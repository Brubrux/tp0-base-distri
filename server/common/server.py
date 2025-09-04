import socket
import logging
import threading
from common import protocol as p
from common import utils as u


class Server:
    def __init__(self, port, listen_backlog, client_number):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

        self._active_connections = []
        self._connections_lock = threading.Lock()
        
        self.agency_status = generate_lottery_diccionary(client_number)
        self.lottery_conducted = False
        self._status_lock = threading.Lock()  # Lock for agency_status and lottery_conducted

        self._file_lock = threading.Lock()    # Lock for utils functions

        self._shutdown_event = threading.Event()

    def run(self):
        """
        Server loop

        Server that accepts new connections and creates a new thread
        for each client connection. Multiple clients can be handled
        simultaneously.
        """
        while not self._shutdown_event.is_set():
            try:
                client_sock = self.__accept_new_connection()
                
                with self._connections_lock:
                    self._active_connections.append(client_sock)
                client_thread = threading.Thread(
                    target=self.__handle_client_connection,
                    args=(client_sock,),
                    daemon=True
                )
                client_thread.start()
                
            except OSError as e:
                if not self._shutdown_event.is_set():
                    logging.error(f'action: accept_connection_error | result: fail | error: {e}')
                else:
                    logging.info(f'action: shutdown_close_socket | result: success')
                break


    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed. This method is thread-safe
        and can handle multiple clients concurrently.
        """
        addr = None
        try:
            addr = client_sock.getpeername()
            logging.info(f'action: start_client_handler | result: success | ip: {addr[0]} | thread: {threading.current_thread().name}')
            
            while not self._shutdown_event.is_set():
                op_code, msg = _recv_message_with_payload_length(client_sock)
                
                if op_code == b'\x02':  # Batch
                    logging.info(f'action: receive_batch | result: success | ip: {addr[0]}')
                    confirmation = self.__handle_bet_batch(msg)
                    response = confirmation.ToBytes()

                elif op_code == b'\x03':  # Get Winners
                    logging.info(f'action: receive_get_winners | result: success | ip: {addr[0]}')
                    response = self.__handle_get_winners(msg).ToBytes()

                elif op_code == b'\x06':  # Agency Ready
                    logging.info(f'action: receive_agency_ready | result: success | ip: {addr[0]}')
                    self.__handle_agency_ready(msg)
                    continue

                elif op_code == b'\xFF':  # Terminate
                    logging.info(f'action: receive_terminate | result: success | ip: {addr[0]}')
                    break
                else:
                    logging.error(f'action: receive_unknown | result: fail | ip: {addr[0]} | error: unknown opcode')
                    raise ValueError("Unknown OpCode")

                _full_send(client_sock, response)
                logging.info(f'action: send_confirmation | result: success | ip: {addr[0]}')

        except ConnectionError as e:
            if addr:
                logging.info(f'action: client_disconnected | result: success | ip: {addr[0]}')
        except OSError as e:
            if addr:
                logging.error(f'action: socket_error | result: fail | ip: {addr[0]} | error: {e}')
        except Exception as e:
            if addr:
                logging.error(f'action: message_processing_error | result: fail | ip: {addr[0]} | error: {e}')
        finally:
            # Remove connection from active connections list
            with self._connections_lock:
                try:
                    self._active_connections.remove(client_sock)
                except ValueError:
                    pass  # already removed
            
            try:
                client_sock.close()
                if addr:
                    logging.info(f'action: close_client_connection | result: success | ip: {addr[0]}')
            except:
                pass
    
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
        1. Set shutdown event
        2. Close listening socket
        3. Close all active connections
        """
        logging.info(f'action: received_SIGTERM | result: in_progress')
        
        self._shutdown_event.set()
        
        # Close listening socket
        try:
            self._server_socket.close()
            logging.info(f'action: close_listening_socket | result: success')
        except Exception as e:
            logging.error(f'action: close_listening_socket | result: fail | error: {e}')

        # Close all connections
        with self._connections_lock:
            connections_count = len(self._active_connections)
            if connections_count > 0:
                logging.info(f'action: closing_active_connections | count: {connections_count}')
                for client_sock in self._active_connections[:]:
                    try:
                        client_sock.close()
                        logging.debug(f'action: close_client_connection | result: success')
                    except Exception as e:
                        logging.warning(f'action: close_client_connection | result: fail | error: {e}')
                
                self._active_connections.clear()
                logging.info(f'action: close_client_connections | result: success | closed_count: {connections_count}')


    def __handle_bet_batch(self, msg):
        try:
            bet_batch = p.BetBatchRegister.DeserializeBetBatch(msg)
        except ValueError as e:
            return p.BetConfirmation(False, "bad_request")

        try:
            with self._file_lock:
                u.store_bets(bet_batch.bets)
        except Exception as e:
            logging.error(f'action: apuesta_recibida | result: fail | cantidad: {len(bet_batch.bets)}')
            return p.BetConfirmation(False, "internal_error")
        
        logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bet_batch.bets)}')
        return p.BetConfirmation(True, f"{len(bet_batch.bets)}")

    def __handle_get_winners(self, msg):
        get_winners_request = p.GetWinners.DeserializeGetWinners(msg)
        logging.info(f'action: decode_get_winners | result: success | agency_id: {get_winners_request.agency_id}')
        
        # Check if lottery has been conducted
        if not self.lottery_ready():
            logging.info(f'action: lottery_status | result: not_conducted | agency_id: {get_winners_request.agency_id}')
            return p.NotConducted()

        winners = self.get_winners(agency_id=get_winners_request.agency_id)
        logging.info(f'action: send_winners | result: success | agency_id: {get_winners_request.agency_id} | winners_count: {len(winners.winner_ids)}')
        return winners

    def __handle_agency_ready(self, msg):
        try:
            agency_ready = p.AgencyReady.Deserialize(msg)
        except ValueError as e:
            return
        # Set ready
        try: 
            with self._status_lock:
                self.agency_status[agency_ready.agency_id] = True
            logging.info(f'action: agency_status_update | result: success | agency_id: {agency_ready.agency_id}')
        except Exception as e:
            logging.error(f'action: agency_status_update | result: fail | error: {e}')

    def get_winners(self, agency_id):
        """
        Get winners for a specific agency
        Thread-safe file access for reading bets
        """
        winners = p.Winners()
        with self._file_lock:
            for b in u.load_bets():
                if b.agency == agency_id and u.has_won(b):
                    winners.add_Id(b.document)
        return winners

    def lottery_ready(self):
        with self._status_lock:
            if not self.lottery_conducted:
                for _, ready in self.agency_status.items():
                    if not ready:
                        return False
                logging.info('action: sorteo | result: success')
                self.lottery_conducted = True

            return True

def generate_lottery_diccionary(client_number):
    d = {}
    for i in range(1, client_number + 1):
        d[i] = False
    return d

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

    # Payload Length
    payload_length_data = _full_recv(sock, 4) 
    payload_length = p.parse_int4_big_endian(payload_length_data)
    
    # Read the exact payload
    payload_data = _full_recv(sock, payload_length)
    return opcode_data, opcode_data + payload_length_data + payload_data