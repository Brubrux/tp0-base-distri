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
        self._active_connection = None

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        while True:
            try:
                client_sock = self.__accept_new_connection()
                self._active_connection = client_sock
                
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
            addr = client_sock.getpeername()
            while True:
                msg = _recv_message_with_payload_length(client_sock)
                if msg is None:
                    logging.info(f'action: receive_terminate | result: success | ip: {addr[0]}')
                    break

                logging.info(f'action: receive_message | result: success | ip: {addr[0]}')

                confirmation = self.__handle_bet_batch(msg)

                logging.info(f'action: send_confirmation | result: in_progress | ip: {addr[0]}')
                
                _full_send(client_sock, confirmation.ToBytes())

                logging.info(f'action: send_confirmation | result: success | ip: {addr[0]}')

        except ConnectionError as e:
            logging.info(f'action: client_disconnected | result: success | ip: {addr[0]}')
        except OSError as e:
            logging.error(f'action: socket_error | result: fail | ip: {addr[0]} | error: {e}')
        except Exception as e:
            logging.error(f'action: message_processing_error | result: fail | ip: {addr[0]} | error: {e}')
        finally:
            self._active_connection = None
            client_sock.close()

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
        Shutdown the server gracefully:
        1. Close listening socket
        2. Close active connection, if there is one
        """
        logging.info(f'action: received_SIGTERM | result: in_progress')
        
        # listening socket
        try:
            self._server_socket.close()
            logging.info(f'action: close_listening_socket | result: success')
        except Exception as e:
            logging.error(f'action: close_listening_socket | result: fail | error: {e}')

        # client connection
        if self._active_connection is not None:
            logging.info(f'action: closing_active_connections | count: 1')
            try:
                self._active_connection.close()
                logging.debug(f'action: close_client_connection | result: success')
            except Exception as e:
                logging.warning(f'action: close_client_connection | result: fail | error: {e}')

            self._active_connection = None
            logging.info(f'action: close_client_connection | result: success | closed_count: 1')


# send and rcv wrappers for handling short-reads/writes

def _full_recv(sock, size):
    """
    Receive exactly 'size' bytes, handling short-reads and connection closures
    """
    data = bytearray()
    while len(data) < size:
        chunk = sock.recv(size - len(data))
        if not chunk:
            # EOF
            raise ConnectionError("Client closed connection")
        data.extend(chunk)
    return bytes(data)

def _full_send(sock, data):
    """
    Send all data, handling short-writes and connection closures
    """
    total_sent = 0
    while total_sent < len(data):
        sent = sock.send(data[total_sent:])
        if sent == 0:
            raise ConnectionError("Socket connection broken during send")
        total_sent += sent

def _recv_message_with_payload_length(sock):
    """
    Receive complete message using payload length field:
    1. Read OpCode (1 byte)
    2. Read Payload Length (4 bytes) 
    3. Read exact payload bytes
    """

    opcode_data = _full_recv(sock, 1)

    if opcode_data == b'\xFF': # Terminate
        logging.info(f'action: receive_terminate | result: in_progress')
        return None

    # Payload Length
    payload_length_data = _full_recv(sock, 4) 
    payload_length = p.parse_int4_big_endian(payload_length_data)
    
    # Read the exact payload
    payload_data = _full_recv(sock, payload_length)
    return opcode_data + payload_length_data + payload_data