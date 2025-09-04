import socket
import logging
from common import protocol as p
from common import utils as u


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        # TODO: Modify this program to handle signal to graceful shutdown
        # the server
        while True:
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
            except OSError as e:
                logging.info(f'action: shutdown_close_socket | result: success')
                break

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            msg = _recv_message_with_payload_length(client_sock)

            addr = client_sock.getpeername()
            logging.info(f'action: receive_message | result: success | ip: {addr[0]}')

            confirmation = self.__handle_bet_batch(msg)

            logging.info(f'action: send_confirmation | result: in_progress | ip: {addr[0]}')
            
            _full_send(client_sock, confirmation.ToBytes())

        except OSError as e:
            logging.error(f'action: receive_message | result: fail | error: {e}')
        except ConnectionError as e:
            logging.error(f'action: connection_error | result: fail | error: {e}')
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c 

    def shutdown(self):
        """
        Shutdown the server
        """
        self._server_socket.close()
        
        logging.info(f'action: received_SIGTERM | result: in_progress')

    def __handle_bet_register(self, msg, addr):
        """
        Handle bet registration message
        """
        try:
            br = p.BetRegister.DeserializeBetRegister(msg)
            logging.info(f'action: decode_bet_register | result: success')
        except Exception as e:
            logging.error(f'action: decode_bet_register | result: fail | error: {e}')
            return p.BetConfirmation(False, "bad_request")
        
        b = u.Bet(
            agency=br.agency_id,
            birthdate=br.birth_date,
            document=br.id,
            first_name=br.first_name,
            last_name=br.last_name,
            number=br.number
        )
        try:
            u.store_bets([b])
        except Exception as e:
            logging.error(f'action: apuesta_almacenada | result: fail | error: {e}')
            return p.BetConfirmation(False, "internal_error")
        logging.info(f'action: apuesta_almacenada | result: success | dni: {br.id} | numero: {br.number}')
        return p.BetConfirmation(True, "")
    
    def __handle_bet_batch(self, msg):
        try:
            bet_batch = p.BetBatchRegister.DeserializeBetBatch(msg)
            logging.info(f'action: decode_bet_batch | result: success | bets_count: {len(bet_batch.bets)}')
        except ValueError as e:
            logging.error(f'action: decode_bet_batch | result: fail | error: {e}')
            return p.BetConfirmation(False, "bad_request")

        try:
            u.store_bets(bet_batch.bets)
        except Exception as e:
            logging.error(f'action: apuesta_recibida | result: fail | cantidad: {len(bet_batch.bets)}')
            return p.BetConfirmation(False, "internal_error")
        logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bet_batch.bets)}')
        return p.BetConfirmation(True, f"{len(bet_batch.bets)}")



# send and rcv wrappers for handling short-reads/writes

def _full_recv(sock, size):
    """
    Receive exactly 'size' bytes, handling short-reads
    """
    data = bytearray()
    while len(data) < size:
        chunk = sock.recv(size - len(data))
        if not chunk:
            raise ConnectionError
        data.extend(chunk)
    return bytes(data)

def _full_send(sock, data):
    """
    Send all data, handling short-writes
    """
    total_sent = 0
    while total_sent < len(data):
        sent = sock.send(data[total_sent:])
        if sent == 0:
            raise ConnectionError
        total_sent += sent

def _recv_message_with_payload_length(sock):
    """
    Receive complete message using payload length field:
    1. Read OpCode (1 byte)
    2. Read Payload Length (4 bytes) 
    3. Read exact payload bytes
    """

    opcode_data = _full_recv(sock, 1)
    
    # Payload Length
    payload_length_data = _full_recv(sock, 4) 
    payload_length = p.parse_int4_big_endian(payload_length_data)
    
    # Read the exact payload
    payload_data = _full_recv(sock, payload_length)
    return opcode_data + payload_length_data + payload_data