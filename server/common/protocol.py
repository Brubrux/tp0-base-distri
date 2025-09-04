import enum

from common import utils


class OpCodes(enum.IntEnum):
    REGISTER = 0x00
    CONFIRM = 0x01
    BATCH = 0X02
    TERMINATE = 0xFF

class BetRegister:

    def __init__(self, agency_id, first_name, last_name, id, birth_date, number):
        self.agency_id = agency_id
        self.first_name = first_name
        self.last_name = last_name
        self.id = id
        self.birth_date = birth_date
        self.number = number

    @staticmethod
    def DeserializeBetRegister(msg):
        index = 0
        # OpCode
        op_code = msg[index]
        if op_code != OpCodes.REGISTER:
            raise ValueError(f"Invalid OpCode: {op_code} should be {OpCodes.REGISTER}")
        index += 1

        # Skip payload length
        index += 4

        # AgencyID
        agency_id = msg[index]
        index += 1

        # FirstName
        first_name_length = msg[index]
        index += 1
        first_name = msg[index:index + first_name_length].decode('utf-8')
        index += first_name_length

        # LastName
        last_name_length = msg[index]
        index += 1
        last_name = msg[index:index + last_name_length].decode('utf-8')
        index += last_name_length

        # ID
        id_length = msg[index]
        index += 1
        id = msg[index:index + id_length].decode('utf-8')
        index += id_length

        # BirthDate
        birth_date_length = msg[index]
        index += 1
        birth_date = msg[index:index + birth_date_length].decode('utf-8')
        index += birth_date_length

        # Number
        number_length = msg[index]
        index += 1
        number = msg[index:index + number_length].decode('utf-8')
        index += number_length

        return BetRegister(agency_id, first_name, last_name, id, birth_date, number)


class BetConfirmation:
    def __init__(self, success, message):
        self.success = success
        self.message = message

    def ToBytes(self):
        message_bytes = self.message.encode('utf-8')
        payload_size = 1 + 1 + len(message_bytes)  # success(1) + messageLength(1) + message
        
        msg = bytearray()
        msg.append(OpCodes.CONFIRM)
        
        msg.extend(payload_size.to_bytes(4, byteorder='big'))
        
        msg.append(self.success)
        msg.append(len(message_bytes))
        msg.extend(message_bytes)
        return bytes(msg)
    

class BetBatchRegister:
    def __init__(self, agency_id, bets, bets_count):
        self.agency_id = agency_id
        self.bets = bets
        self.bets_count = bets_count

    @staticmethod
    def DeserializeBetBatch(msg):
        index = 0
        # OpCode
        op_code = msg[index]
        if op_code != OpCodes.BATCH:
            raise ValueError(f"Invalid OpCode: {op_code} should be {OpCodes.BATCH}")
        index += 1

        # Skip payload length
        index += 4

        # AgencyID
        agency_id = msg[index]
        index += 1

        # Bets count 4B
        bets_count = parse_int4_big_endian(msg[index:index + 4])
        index += 4

        # Bets -> 1b string len, "name,lastname,..."
        bets = []
        for _ in range(bets_count):
            bet_length = msg[index]
            index += 1
            bet_csv = msg[index:index + bet_length].decode('utf-8')
            index += bet_length
            bets.append(csv_to_bet(bet_csv, agency_id))

        if bets_count != len(bets):
            raise ValueError(f"Invalid bets count: read {len(bets)}, expected {bets_count}")
        
        return BetBatchRegister(agency_id, bets, bets_count)


def csv_to_bet(csv_string, agency_id):
    """
    Converts a CSV string to a Bet object from module utils.
    String format expected: "first_name,last_name,id,birth_date,number"
    """
    fields = csv_string.split(',')
    if len(fields) != 5:
        raise ValueError(f"Invalid CSV format: {csv_string}")
   
    return utils.Bet(
        agency=agency_id,
        first_name=fields[0],
        last_name=fields[1],
        document=fields[2],
        birthdate=fields[3],
        number=fields[4]
    )

def parse_int4_big_endian(data):
    if len(data) != 4:
        raise ValueError("Data must be exactly 4 bytes long")
    return (data[0] << 24) + (data[1] << 16) + (data[2] << 8) + data[3]