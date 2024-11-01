# TCP server with PoW protection from DDOS

## 1. Task description
Design and implement some tcp server.

- TCP server should be protected from DDOS attacks with the Prof of Work
(https://en.wikipedia.org/wiki/Proof_of_work), the challenge-response protocol should be used.
- The choice of the PoW algorithm should be explained.
- After Prof Of Work verification, server should send one of the quotes from "word of
wisdom" book or any other collection of the quotes.
- Docker file should be provided both for the server and for the client that solves the
PoW challenge.

## 2. Getting started

### 2.1 Makefile commands
| Command      | Description           |
|--------------|-----------------------|
| help         | Show command list     |
| lint         | Run code linters      |
| test         | Run unit-tests        |
| mock         | Generate mocks        |
| build_server | Build a binary server |
| build_client | Build a binary client |
| startd       | Compose run detached  |
| start        | Compose run           |
| stop         | Compose down          |
| log          | Compose logs          |
| logserver    | Compose server logs   |
| logclient    | Compose client logs   |


### 2.2 Start server and client
```
make start
```

## 3 Protocol definition

### 3.1 Request definition
| RequestType | ResourceID | ContentLength |  Payload   |
|:-----------:|:----------:|:-------------:|:----------:|
|  [1]bytes   |  [2]bytes  |   [4]bytes    | [...]bytes |

#### 3.1.1 RequestType
Request type can take one of the following values:
- `0x00`: RequestTypeExit      
- `0x01`: RequestTypeChallenge
- `0x02`: RequestTypeResource

#### 3.1.2 ResourceID
An arbitrary resource identifier. Needed to determine which handler was called.

#### 3.1.3 ContentLength
Request body length (Payload)

#### 3.1.4 Payload
Optional request body

### 3.2 Response definition
|  Status  | ContentLength |  Payload   |
|:--------:|:-------------:|:----------:|
| [1]bytes |   [4]bytes    | [...]bytes |

#### 3.2.1 Status
Response status can take one of the following values:
- `0x00`: StatusOK
- `0x01`: StatusErr

#### 3.2.2 ContentLength
Response body length (Payload)

#### 3.2.3 Payload
Optional response body

## 4. Proof of Work
The idea of Proof of Work for DDOS protection is that a client who wants to get some resource from the server must first solve some problem issued by the server. 
To complete the task, the client's computing resources are required, which means that implementing a DDOS attack becomes more expensive for the attacker. 
At the same time, on the server side, checking the result of the task performed by the client requires practically no resources.

### 4.1 Selection of an algorithm
I looked at several existing algorithms:
- [Merkle tree](https://en.wikipedia.org/wiki/Merkle_tree)
- [Hashcash](https://en.wikipedia.org/wiki/Hashcash)

Disadvantages of current algorithms: 
- In a Merkle tree, the server has to do too much work to verify the client's solution. Checking the solution is not possible in O(1)
- In Hashcash, the client can make calculations in advance and use them for queries. You can combat this by storing tasks issued to clients on the server side.

Based on the shortcomings of existing algorithms, I chose to implement my own based on Hashcash.
The algorithm is based on two things:
- Signing the issued task with a secret key. This makes it difficult for the client to precompute tasks.
- Instead of explicitly looking for leading zeros in the hash, a bitwise shift and big.Int comparison are used to check the solution.

However, despite the signature, the client can continue to reuse the solution within another connection.
Therefore, you should not indicate too much time for the client to solve the problem.
In a production environment, this drawback can be eliminated by using a centralized cache.

The server optionally accepts in-memory cache. 
If the cache is nil, the server will not be able to check that it has already checked the Challange due to the request.
This potentially makes it possible to send the same result of the completed 
Challenge several times during its lifespan.

Structure used for PoW:
```go
type Challenge struct {
	Signature     []byte `json:"sig"`
	UnixTimestamp int64  `json:"unix"`
	Nonce         int64  `json:"nonce"`
	Rand          []byte `json:"rand"`
	Difficulty    uint8  `json:"dif"`
}
```

Task verification algorithm complexity O(1), and works as follows:
1. Select the difficulty of the task in the range from 0 to 255, for example 20
2. Shift one byte to the left, 1 << 255-20. The lower the complexity is selected, the smaller the shift will be and the larger the resulting value will be.
3. At the output we get a large number represented by big.Int
4. Calculate sha256 for Rand and Nonce
5. Big.Int is created from the resulting hash amount
6. The resulting big.Int must be less than the number from point 3


## 5. Structure of folders
Existing directories:
+ client - client package for external usage
+ cmd/server - main.go for server
+ cmd/client - main.go for client
+ internal/config - config files for server and client
+ internal/handler - example handler
+ internal/pow - logic of PoW
+ internal/proto - protocol implementation
+ internal/server - tcp server implementation

## 6. Possible improves
- Dynamic difficulty change.
- Integrations tests.
- Implement the ability to support handlers with request parameters.
- Configuring client and server using options pattern