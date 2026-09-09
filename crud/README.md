# API CRUD em Go — Arquitetura em Camadas (Material de Estudo)

> Guia de estudo montado a partir de uma sessão de perguntas e respostas.
> Cobre a construção de uma API de `customers` em Go usando **apenas a biblioteca padrão** (`net/http` do Go 1.22+), separada em camadas: **repository → service → handler**.

---

## Índice

1. [Visão geral da arquitetura](#visão-geral-da-arquitetura)
2. [Estrutura de pastas](#estrutura-de-pastas)
3. [Passo 1 — Model](#passo-1--o-model)
4. [Passo 2 — Repository](#passo-2--o-repository)
5. [Passo 3 — Service](#passo-3--o-service)
6. [Passo 4 — Handler](#passo-4--o-handler)
7. [Passo 5 — Helpers de resposta](#passo-5--helpers-de-resposta)
8. [Passo 6 — main.go (montagem)](#passo-6--maingo--onde-tudo-se-conecta)
9. [Como rodar](#como-rodar)
10. [Conceitos-chave](#conceitos-chave-o-porquê-de-cada-coisa)

---

## Visão geral da arquitetura

A API é dividida em três camadas, cada uma com uma responsabilidade única:

```
   Requisição HTTP
        │
        ▼
┌───────────────┐   traduz HTTP ↔ struct
│    HANDLER    │   (parse do request, status codes, JSON)
└───────┬───────┘
        │ chama
        ▼
┌───────────────┐   regra de negócio
│    SERVICE    │   (validações, decisões)
└───────┬───────┘
        │ chama
        ▼
┌───────────────┐   acesso a dados
│  REPOSITORY   │   (banco, cache, ou map em memória)
└───────┬───────┘
        │
        ▼
      Dados
```

**Regra de ouro:** cada camada só conhece a de baixo, e **através de uma interface**.
Isso é o que torna tudo testável e trocável (trocar o `map` por Postgres, por exemplo,
sem tocar em service nem handler).

---

## Estrutura de pastas

```
customers-api/
├── cmd/api/main.go                              # ponto de entrada
├── internal/
│   ├── model/customer.go                        # entidade
│   ├── repository/customer_repository.go        # acesso a dados
│   ├── service/customer_service.go              # regra de negócio
│   ├── handler/customer_handler.go              # camada HTTP
│   └── handler/response.go                       # helpers de resposta
└── go.mod
```

> **Sobre `internal/`:** é uma pasta especial reconhecida pelo compilador do Go.
> Qualquer código dentro dela só pode ser importado por pacotes do mesmo projeto,
> protegendo o código interno de ser usado por projetos externos.

---

## Passo 1 — O Model

`internal/model/customer.go`

A entidade central. Fica sozinha porque **todas** as camadas dependem dela.

```go
package model

type Customer struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
```

> As tags `` `json:"..."` `` definem como cada campo do struct vira JSON
> (e como o JSON recebido é mapeado de volta para o struct).

---

## Passo 2 — O Repository

`internal/repository/customer_repository.go`

Só cuida de **guardar e buscar dados**. Nada de HTTP, nada de regra de negócio.
Usa um `map` em memória para rodar sem banco. Quando trocar por Postgres, só o
*interior* destes métodos muda; o resto do projeto nem fica sabendo.

```go
package repository

import (
	"errors"
	"sync"

	"meu-projeto/internal/model"
)

// erro de domínio que as camadas de cima podem checar
var ErrNotFound = errors.New("cliente não encontrado")

type CustomerRepository struct {
	mu        sync.Mutex              // protege o map de acessos concorrentes
	customers map[int]model.Customer  // "tabela" em memória
	nextID    int
}

func NewCustomerRepository() *CustomerRepository {
	return &CustomerRepository{
		customers: make(map[int]model.Customer),
		nextID:    1,
	}
}

func (r *CustomerRepository) FindAll() ([]model.Customer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	list := make([]model.Customer, 0, len(r.customers))
	for _, c := range r.customers {
		list = append(list, c)
	}
	return list, nil
}

func (r *CustomerRepository) FindByID(id int) (model.Customer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.customers[id]
	if !ok {
		return model.Customer{}, ErrNotFound
	}
	return c, nil
}

func (r *CustomerRepository) Create(c model.Customer) (model.Customer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c.ID = r.nextID
	r.customers[c.ID] = c
	r.nextID++
	return c, nil
}

func (r *CustomerRepository) Update(id int, c model.Customer) (model.Customer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.customers[id]; !ok {
		return model.Customer{}, ErrNotFound
	}
	c.ID = id
	r.customers[id] = c
	return c, nil
}

func (r *CustomerRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.customers[id]; !ok {
		return ErrNotFound
	}
	delete(r.customers, id)
	return nil
}
```

### Sobre `mu sync.Mutex` e `customers map[int]model.Customer`

Esses dois campos trabalham juntos:

- **`customers map[int]model.Customer`** é o **armazenamento**: um mapa (dicionário)
  chave→valor. A chave é o `id` (`int`) e o valor é o `Customer` inteiro. Faz o papel
  do banco de dados no exemplo.

- **`mu sync.Mutex`** é o **porteiro** que protege o map. Quando o servidor recebe
  várias requisições ao mesmo tempo, o Go atende cada uma em paralelo (em *goroutines*).
  **Map em Go não é seguro para acesso concorrente** — dois writes simultâneos
  fazem o programa *crashar* (`fatal error: concurrent map writes`). O mutex
  (*mutual exclusion*) funciona como uma **chave de banheiro**: só uma goroutine entra
  por vez. `Lock()` pega a chave, `Unlock()` devolve.

```go
r.mu.Lock()          // pega a chave — daqui pra frente, sou só eu
defer r.mu.Unlock()  // devolve a chave quando a função terminar
r.customers[c.ID] = c   // mexe no map com segurança
```

> **Importante:** o mutex só é necessário porque guardamos em memória. Ao trocar
> por Postgres, ele **some** — o banco já cuida da concorrência sozinho.

---

## Passo 3 — O Service

`internal/service/customer_service.go`

Aqui mora a **regra de negócio** (validações, decisões). O ponto mais importante da
arquitetura aparece aqui: o service **não depende do repository concreto, e sim de uma
interface**. É isso que permite trocar o `map` por Postgres — ou por um mock nos testes —
sem tocar no service.

```go
package service

import (
	"errors"

	"meu-projeto/internal/model"
)

// O service define a interface que ELE precisa.
// Qualquer repository que tenha esses métodos serve.
type CustomerRepository interface {
	FindAll() ([]model.Customer, error)
	FindByID(id int) (model.Customer, error)
	Create(c model.Customer) (model.Customer, error)
	Update(id int, c model.Customer) (model.Customer, error)
	Delete(id int) error
}

var ErrInvalidInput = errors.New("name e email são obrigatórios")

type CustomerService struct {
	repo CustomerRepository // depende da INTERFACE, não do struct concreto
}

func NewCustomerService(repo CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) List() ([]model.Customer, error) {
	return s.repo.FindAll()
}

func (s *CustomerService) GetByID(id int) (model.Customer, error) {
	return s.repo.FindByID(id)
}

func (s *CustomerService) Create(c model.Customer) (model.Customer, error) {
	if c.Name == "" || c.Email == "" {
		return model.Customer{}, ErrInvalidInput
	}
	return s.repo.Create(c)
}

func (s *CustomerService) Update(id int, c model.Customer) (model.Customer, error) {
	if c.Name == "" || c.Email == "" {
		return model.Customer{}, ErrInvalidInput
	}
	return s.repo.Update(id, c)
}

func (s *CustomerService) Delete(id int) error {
	return s.repo.Delete(id)
}
```

---

## Passo 4 — O Handler

`internal/handler/customer_handler.go`

Fica **enxuto**: só traduz HTTP ↔ struct e chama o service. Também depende de uma
**interface de service**, não do concreto.

```go
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"meu-projeto/internal/model"
	"meu-projeto/internal/repository"
	"meu-projeto/internal/service"
)

// interface que o handler precisa do service
type CustomerService interface {
	List() ([]model.Customer, error)
	GetByID(id int) (model.Customer, error)
	Create(c model.Customer) (model.Customer, error)
	Update(id int, c model.Customer) (model.Customer, error)
	Delete(id int) error
}

type CustomerHandler struct {
	service CustomerService
}

func NewCustomerHandler(s CustomerService) *CustomerHandler {
	return &CustomerHandler{service: s}
}

// Cada handler registra as PRÓPRIAS rotas.
// Assim o router nunca cresce: ele só chama RegisterRoutes de cada handler.
func (h *CustomerHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /customers", h.List)
	mux.HandleFunc("POST /customers", h.Create)
	mux.HandleFunc("GET /customers/{id}", h.GetByID)
	mux.HandleFunc("PUT /customers/{id}", h.Update)
	mux.HandleFunc("DELETE /customers/{id}", h.Delete)
}

func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input model.Customer
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	created, err := h.service.Create(input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao criar")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *CustomerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	c, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	var input model.Customer
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	updated, err := h.service.Update(id, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, repository.ErrNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "erro ao atualizar")
		}
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *CustomerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	if err := h.service.Delete(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao deletar")
		return
	}
	w.WriteHeader(http.StatusNoContent) // 204, sem corpo
}

func idFromPath(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}
```

### Roteamento no `net/http` (Go 1.22+)

- O método vem junto no padrão: `"GET /customers"` (método em maiúsculo + espaço + path).
- `{id}` captura um trecho da URL; leia com `r.PathValue("id")`.
- Não há agrupamento por prefixo (você repete `/customers` em cada linha) nem `r.Use()`
  para middleware de grupo — é o preço de não ter dependências externas.

---

## Passo 5 — Helpers de resposta

`internal/handler/response.go`

Separados num arquivo próprio porque **todos** os handlers vão reutilizar.
Como são do mesmo pacote `handler`, o `customer_handler.go` os enxerga sem import.

```go
package handler

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
```

> **Curiosidade do Go:** os métodos/funções de um pacote podem estar espalhados
> em arquivos diferentes. O compilador junta tudo do mesmo `package`.

---

## Passo 6 — main.go — onde tudo se conecta

`cmd/api/main.go`

Aqui a **injeção de dependência** fica visível: repository entra no service,
service entra no handler. É a "montagem" da aplicação, de baixo pra cima.

```go
package main

import (
	"log"
	"net/http"

	"meu-projeto/internal/handler"
	"meu-projeto/internal/repository"
	"meu-projeto/internal/service"
)

func main() {
	mux := http.NewServeMux()

	// montagem das camadas: repo → service → handler
	customerRepo := repository.NewCustomerRepository()
	customerSvc := service.NewCustomerService(customerRepo)
	customerHandler := handler.NewCustomerHandler(customerSvc)

	customerHandler.RegisterRoutes(mux)

	log.Println("rodando em :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

---

## Como rodar

```bash
# na raiz do projeto, só na primeira vez:
go mod init meu-projeto     # cria o go.mod

# sobe o servidor em :8080
go run ./cmd/api
```

> **Atenção ao nome do módulo:** o prefixo dos imports (`meu-projeto/internal/...`)
> tem que bater **exatamente** com o `module` declarado no `go.mod`. Se usar outro
> nome (ex.: `github.com/seuusuario/customers-api`), troque em **todos** os imports.
> Esse é o erro nº 1 de "package not found" em quem está começando.

O `go.mod` fica assim:

```
module meu-projeto

go 1.23
```

### Testando os endpoints

```bash
# criar
curl -X POST localhost:8080/customers -d '{"name":"Ana","email":"ana@x.com"}'

# listar
curl localhost:8080/customers

# buscar por id
curl localhost:8080/customers/1

# atualizar
curl -X PUT localhost:8080/customers/1 -d '{"name":"Ana Silva","email":"ana@x.com"}'

# deletar
curl -X DELETE localhost:8080/customers/1
```

---

## Conceitos-chave (o porquê de cada coisa)

### Receiver — `func (h *CustomerHandler) Metodo(...)`

O trecho `(h *CustomerHandler)` entre `func` e o nome do método é o **receiver**.
Ele "prende" a função ao tipo, transformando-a num método. Dentro do corpo, `h` é a
instância (equivale ao `this`/`self` de outras linguagens), mas é só um nome de variável
que você escolhe.

- Em Go **não existe classe**: o tipo é definido no `struct` e os métodos ficam soltos
  no pacote, ligados pelo receiver.
- O `*` (`*CustomerHandler`) torna o receiver um **ponteiro**: o método acessa o struct
  original (pode modificá-lo) e evita copiar o struct a cada chamada.
- **Regra de bolso:** se algum método do tipo precisa de ponteiro, use ponteiro em todos.

### Interface implícita

Em Go, uma interface é um **contrato**. Um tipo satisfaz a interface automaticamente se
tiver os métodos com a **assinatura exata** — sem declarar `implements` em lugar nenhum.

```go
type Registerer interface {
	RegisterRoutes(mux *http.ServeMux)
}
```

Qualquer handler com `RegisterRoutes(*http.ServeMux)` já é um `Registerer`.
"Assinatura" inclui: **nome** do método, **tipos** dos parâmetros (não os nomes das
variáveis) e o **retorno**. Mudou qualquer um → não satisfaz mais.

### Variadic — `handlers ...Registerer`

O `...` significa "zero ou mais argumentos desse tipo". Dentro da função vira um slice.

```go
func New(handlers ...Registerer) *http.ServeMux {
	for _, h := range handlers {   // percorre como lista
		h.RegisterRoutes(mux)
	}
}
```

Regras: só um parâmetro variadic, e sempre o último. Para passar um slice pronto,
use `...` na chamada: `New(lista...)`.

### `mux` e `*http.ServeMux`

- **`mux`** = *multiplexer*. É o **roteador**: recebe muitas requisições diferentes e
  decide qual função atende cada uma. `mux.HandleFunc("GET /users", h.List)` ensina:
  "quando chegar `GET /users`, mande pro `h.List`".
- **`http.ServeMux`** é o tipo concreto (o multiplexador padrão do `net/http`).
- **`*http.ServeMux`** é um **ponteiro** para ele — necessário porque registrar rota
  *modifica* o mux. Cria-se com `http.NewServeMux()`.
- O `*http.ServeMux` **é** um `http.Handler` (tem o método `ServeHTTP`), por isso pode
  ser passado direto ao `http.ListenAndServe`.

### Injeção de dependência (DI)

**Conceito:** entregar a dependência **de fora** (por parâmetro) em vez de o objeto
criá-la sozinho lá dentro.

```go
// SEM injeção — o service cria o repo escondido dentro dele
func NewService() *Service {
	return &Service{repo: NewRepository()}
}

// COM injeção — o repo chega pronto, de fora
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}
```

**Por que importa:** permite trocar a dependência (Postgres em produção, um fake no teste)
sem alterar quem a usa.

**No Go:** DI existe e é **manual e explícita** — só passar argumentos, geralmente no
`main.go`. Não há framework nem "container mágico" como o Spring (Java). O poder real
aparece ao injetar uma **interface** em vez do struct concreto.

```go
// produção
svc := NewCustomerService(NewCustomerRepository())
// teste
svc := NewCustomerService(&fakeRepo{})  // um repo de mentira que cumpre a interface
```

### Padrões idiomáticos que aparecem no código

- **`c, ok := mapa[chave]`** — o segundo valor (`ok`, booleano) diz se a chave existe.
  É como distinguir "não encontrado" (404) de um registro real.
- **`defer`** — agenda algo para rodar quando a função terminar, por qualquer caminho
  de saída. Usado para `Unlock()` de mutex, fechar conexões/arquivos, etc.
- **`errors.Is(err, ErrAlgumaCoisa)`** — checa *qual* erro veio de baixo. O erro "sobe"
  pelas camadas carregando o significado; só o handler o traduz para status HTTP.
- **`json.NewDecoder(r.Body).Decode(&input)`** — lê o corpo JSON para o struct
  (passa `&input`, o ponteiro, porque o Decode precisa *modificar* o struct).
- **`json.NewEncoder(w).Encode(data)`** — faz o caminho inverso: struct → JSON na resposta.

---

## Próximos passos sugeridos

1. **Trocar o `map` por Postgres:** crie `repository/customer_postgres.go` com um struct
   que tenha os mesmos 5 métodos (`FindAll`, `FindByID`, `Create`, `Update`, `Delete`)
   rodando SQL. Como satisfaz a mesma interface, a **única** linha que muda no projeto é
   no `main.go`. Service e handler ficam intactos.

2. **Adicionar uma segunda entidade** (ex.: `products`) seguindo o mesmo padrão, para ver
   a arquitetura escalar: cada domínio autocontido, com seu próprio `RegisterRoutes`.

3. **Escrever um teste do service** com um repository fake, sem banco, para sentir na
   prática o retorno da injeção de dependência + interface.

4. **GraphQL como segunda "porta de entrada":** REST e GraphQL podem coexistir sobre o
   mesmo service/repository. No GraphQL, uma única rota (`POST /graphql`) atende tudo;
   `query` lê dados (equivale ao GET) e `mutation` cria/atualiza/deleta (equivale a
   POST/PUT/DELETE). Cada *resolver* chama o mesmo `CustomerService`.
