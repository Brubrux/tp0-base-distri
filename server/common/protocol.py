import enum


class OpCodes(enum.IntEnum):
    REGISTER = 0x00
    CONFIRM = 0x01


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
        msg = bytearray()
        msg.append(OpCodes.CONFIRM)
        msg.append(self.success)
        msg.append(len(self.message))
        msg.extend(self.message.encode('utf-8'))
        return bytes(msg)