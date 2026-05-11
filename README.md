# distributed-auth

Serviço de autenticação distribuído em Go com load balancer, suporte a gRPC e RabbitMQ, autenticação via JWT e persistência em PostgreSQL.

O projeto foi construído como um estudo prático de sistemas distribuídos, comparando duas abordagens de transporte para o pipeline de autenticação: **gRPC (Protocol Buffers)** e **RabbitMQ (AMQP)**, com benchmarks de throughput e latência entre as duas.

---

## Funcionalidades

- Registro e login de usuários com hash de senha via **bcrypt**
- Geração e validação de tokens **JWT**
- Load balancer simples distribuindo requisições entre múltiplas instâncias do servidor
- Suporte a dois modos de transporte: **gRPC** e **RabbitMQ**
- Persistência de usuários em **PostgreSQL**
- Benchmark de carga para comparar desempenho entre os transportes

---

## Tecnologias

| Camada | Tecnologia |
|---|---|
| Linguagem | Go 1.24 |
| Transporte | gRPC / RabbitMQ (AMQP) |
| Autenticação | JWT (`golang-jwt/jwt`) |
| Criptografia | bcrypt (`golang.org/x/crypto`) |
| Banco de dados | PostgreSQL (`lib/pq`) |
| Serialização | Protocol Buffers |
| Infraestrutura | Docker / Docker Compose |

---

## Estrutura do Projeto

```
distributed-auth/
├── cmd/
│   ├── server/        # Entrypoint do servidor gRPC
│   └── client/        # Cliente de testes e benchmark
├── deployments/       # docker-compose (PostgreSQL)
├── internal/          # Lógica de negócio (auth, db, handlers)
├── pkg/config/        # Configurações da aplicação
├── proto/             # Definição do contrato gRPC (auth.proto)
├── scripts/           # Scripts auxiliares
├── Makefile
├── go.mod
└── go.sum
```

---

## Pré-requisitos

- [Go 1.24+](https://golang.org/dl/)
- [Docker e Docker Compose](https://docs.docker.com/get-docker/)
- [protoc](https://grpc.io/docs/protoc-installation/) + plugins Go (somente se for regenerar o código protobuf)

---

## Como rodar

### 1. Suba o PostgreSQL

```bash
make docker-up
```

### 2. Inicie um servidor gRPC

```bash
make run-server
```

Para subir 3 servidores em paralelo (portas 9001, 9002 e 9003):

```bash
make run-servers
```

### 3. Execute o cliente de testes

```bash
make run-client
```

---

## Comandos disponíveis

```bash
make proto        # Gera código Go a partir do auth.proto
make run-server   # Sobe um servidor gRPC na porta 9001
make run-servers  # Sobe 3 servidores gRPC (9001–9003)
make run-client   # Executa os testes e benchmarks do cliente
make test         # Roda os testes unitários com cobertura
make build        # Compila os binários em bin/
make docker-up    # Sobe o PostgreSQL via Docker Compose
make docker-down  # Para os serviços Docker
make clean        # Remove binários e cache
make help         # Exibe todos os comandos disponíveis
```

---

## Variáveis de ambiente

| Variável | Descrição | Padrão |
|---|---|---|
| `GRPC_PORT` | Porta do servidor gRPC | `9001` |
| `SERVER_ID` | Identificador da instância | `grpc-server-1` |

---

## Motivação

Este projeto foi desenvolvido para explorar na prática conceitos de sistemas distribuídos:

- Como um load balancer simples distribui carga entre instâncias stateless
- Trade-offs de latência e throughput entre gRPC (síncrono, binário) e RabbitMQ (assíncrono, mensageria)
- Autenticação segura com JWT e bcrypt em um ambiente multi-instância

---

## Autor

**Renato Moreira Serrano de Andrade**  
[github.com/renatomsa](https://github.com/renatomsa) · [linkedin.com/in/renato-serrano-269a4a210](https://linkedin.com/in/renato-serrano-269a4a210)
