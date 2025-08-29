import sys 

NAME = "name: tp0"

SERVICES = "services:"

SERVER = """  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    volumes:
      - ./server/config.ini:/server/config.ini
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net
"""
CLIENT = """  client1:
    container_name: client1
    image: client:latest
    entrypoint: /client
    volumes:
      - ./client/config.yaml:/client/config.yaml
    environment:
      - CLI_ID=1
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server
"""

NETWORKS = """networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"""

def client_maker(client_id):
    return f"""  client{client_id}:
    container_name: client{client_id}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID={client_id}
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server
"""

def compose_maker(file_name, client_num):
    with open(file_name, "w") as f:
        f.write(NAME + '\n')
        f.write(SERVICES + '\n')
        f.write(SERVER + '\n')
        for i in range(1, client_num+1):
            f.write(client_maker(i))
        f.write(NETWORKS)

def main(args):
    if len(args) != 3: 
        print(f"Invalid args lenght. Expected 2 got {max(0, len(args)-1)}")
    file_name, client_num = args[1], args[2]
    compose_maker(file_name, int(client_num))

if __name__ == "__main__":
    main(sys.argv)