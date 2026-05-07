# Rinha de Backend 2026 — Go

Eu sou o **Nithack** e esta é a minha implementação em **Go** para a **Rinha de Backend 2026**.

A proposta deste projeto é resolver o desafio de detecção de fraude mantendo o escopo bem fechado no enunciado: receber uma transação, transformá-la no vetor oficial de 14 dimensões, comparar esse vetor com a base de referências, buscar os 5 vizinhos mais próximos e responder se a transação deve ser aprovada ou não.

A minha prioridade aqui foi simplicidade operacional, previsibilidade e aderência ao desafio. Não adicionei banco de dados, fila, cache externo, heurística extra de fraude ou regra que não esteja relacionada diretamente ao problema proposto.

---

## Objetivo

A aplicação expõe dois endpoints públicos através do load balancer:

GET /ready
POST /fraud-score

O endpoint principal é:

POST /fraud-score

Ele recebe o payload da transação, calcula o vetor de características e retorna:

{
  "approved": true,
  "fraud_score": 0.2
}

A decisão final segue a regra do desafio:

fraud_score = quantidade_de_fraudes_entre_os_5_vizinhos / 5

approved = fraud_score < 0.6

Ou seja:

Fraudes entre os 5 vizinhos	fraud_score	approved

0	0.0	true
1	0.2	true
2	0.4	true
3	0.6	false
4	0.8	false
5	1.0	false



---

Stack

Usei uma stack bem direta:

Go
Nginx
Docker
Docker Compose

A arquitetura tem:

Nginx
 ├── api1
 └── api2
       └── vector-engine

O Nginx recebe as requisições na porta 9999 e distribui entre duas instâncias da API.

As APIs não carregam o dataset completo de referências. Elas apenas:

1. recebem o payload;


2. carregam normalization.json;


3. carregam mcc_risk.json;


4. geram o vetor oficial de 14 dimensões;


5. enviam esse vetor para o vector-engine;


6. calculam fraud_score;


7. retornam a resposta final.



O vector-engine é um serviço interno. Ele carrega os vetores de referência pré-processados em memória e executa a busca dos 5 vizinhos mais próximos.


---

Estrutura do projeto

.
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── nginx.conf
├── resources/
│   ├── references.json.gz
│   ├── mcc_risk.json
│   └── normalization.json
├── cmd/
│   ├── api/
│   │   └── main.go
│   ├── preprocess/
│   │   └── main.go
│   └── vector-engine/
│       └── main.go
└── internal/
    └── challenge/
        ├── constants.go
        ├── files.go
        ├── types.go
        └── vectorize.go


---

Arquivos oficiais esperados

Antes de buildar o projeto, os arquivos oficiais do desafio precisam estar em:

resources/references.json.gz
resources/mcc_risk.json
resources/normalization.json

Esses arquivos são usados assim:

Arquivo	Uso

references.json.gz	Base de vetores de referência usada na busca dos vizinhos mais próximos
mcc_risk.json	Mapa de risco por MCC
normalization.json	Valores máximos usados para normalizar os campos numéricos



---

Pré-processamento

Eu pré-processo o arquivo references.json.gz durante o build da imagem do vector-engine.

O motivo é simples: não faz sentido pagar o custo de gzip + JSON no runtime se o dataset oficial não muda durante o teste.

O pré-processador gera:

/app/processed/references.f32
/app/processed/labels.bin

references.f32

Contém todos os vetores em formato binário:

14 float32 por registro
little-endian

labels.bin

Contém o label de cada vetor:

0 = legit
1 = fraud

Com isso, o vector-engine sobe mais simples: ele apenas carrega bytes já prontos para busca.


---

Vetorização

A API transforma cada transação recebida em um vetor de 14 dimensões.

A ordem seguida é:

0  amount
1  installments
2  amount_vs_avg
3  hour_of_day
4  day_of_week
5  minutes_since_last_tx
6  km_from_last_tx
7  km_from_home
8  tx_count_24h
9  is_online
10 card_present
11 unknown_merchant
12 mcc_risk
13 merchant_avg_amount

Algumas regras importantes:

last_transaction = null

Quando não existe última transação, uso o sentinela oficial:

minutes_since_last_tx = -1
km_from_last_tx = -1

MCC desconhecido

Quando o MCC não existe em mcc_risk.json, uso:

mcc_risk = 0.5

Dia da semana

O dia da semana é normalizado com segunda-feira como 0 e domingo como 6.

segunda = 0
terça   = 1
quarta  = 2
quinta  = 3
sexta   = 4
sábado  = 5
domingo = 6

Depois disso:

day_of_week = dia / 6

Hora do dia

A hora é normalizada assim:

hour_of_day = hora / 23


---

Busca vetorial

O vector-engine recebe um vetor binário interno com 14 float32.

Ele executa a busca dos vizinhos mais próximos usando distância euclidiana quadrática:

distance = soma((referencia[i] - consulta[i])²)

Não calculo raiz quadrada porque ela não muda a ordenação das distâncias. Para comparar quem está mais perto, a distância quadrática é suficiente.

Durante a busca, mantenho apenas os 5 melhores candidatos.

O serviço retorna apenas um byte para a API:

quantidade de fraudes entre os 5 vizinhos

Exemplo:

0, 1, 2, 3, 4 ou 5

A API transforma isso em fraud_score.


---

Endpoints

GET /ready

Verifica se a aplicação está pronta.

A API chama internamente o /ready do vector-engine.

Resposta esperada:

204 No Content


---

POST /fraud-score

Recebe uma transação e retorna a decisão.

Exemplo de request:

{
  "id": "tx-1329056812",
  "transaction": {
    "amount": 41.12,
    "installments": 2,
    "requested_at": "2026-03-11T18:45:53Z"
  },
  "customer": {
    "avg_amount": 82.24,
    "tx_count_24h": 3,
    "known_merchants": ["MERC-003", "MERC-016"]
  },
  "merchant": {
    "id": "MERC-016",
    "mcc": "5411",
    "avg_amount": 60.25
  },
  "terminal": {
    "is_online": false,
    "card_present": true,
    "km_from_home": 29.23
  },
  "last_transaction": null
}

Exemplo de response:

{
  "approved": true,
  "fraud_score": 0.2
}


---

Como rodar localmente

1. Colocar os arquivos oficiais

Crie a pasta resources e coloque os arquivos do desafio:

mkdir -p resources

Estrutura esperada:

resources/references.json.gz
resources/mcc_risk.json
resources/normalization.json


---

2. Buildar as imagens

docker build --target api -t nithack/rinha-2026-api:latest .
docker build --target vector-engine -t nithack/rinha-2026-vector-engine:latest .

Durante o build da imagem vector-engine, o pré-processamento será executado automaticamente.


---

3. Subir os containers

docker compose up -d


---

4. Verificar se está pronto

curl -i http://localhost:9999/ready

Resposta esperada:

HTTP/1.1 204 No Content


---

5. Testar o score

curl -s http://localhost:9999/fraud-score \
  -H 'Content-Type: application/json' \
  -d '{
    "id": "tx-1329056812",
    "transaction": {
      "amount": 41.12,
      "installments": 2,
      "requested_at": "2026-03-11T18:45:53Z"
    },
    "customer": {
      "avg_amount": 82.24,
      "tx_count_24h": 3,
      "known_merchants": ["MERC-003", "MERC-016"]
    },
    "merchant": {
      "id": "MERC-016",
      "mcc": "5411",
      "avg_amount": 60.25
    },
    "terminal": {
      "is_online": false,
      "card_present": true,
      "km_from_home": 29.23
    },
    "last_transaction": null
  }'


---

Recursos definidos no Docker Compose

A distribuição de recursos foi pensada para manter o maior peso no vector-engine, porque ele é quem carrega o dataset e executa a busca.

lb:
  cpus: "0.05"
  memory: "16MB"

api1:
  cpus: "0.10"
  memory: "32MB"

api2:
  cpus: "0.10"
  memory: "32MB"

vector-engine:
  cpus: "0.75"
  memory: "270MB"

Total:

CPU: 1.00
Memória: 350MB


---

Decisões de implementação

Por que separar API e vector-engine?

Porque o desafio exige duas instâncias de API. Se cada API carregasse o dataset completo, eu duplicaria memória sem necessidade.

A separação permite:

api1 leve
api2 leve
vector-engine centralizado com o dataset

Isso reduz duplicação de memória e deixa a responsabilidade mais clara.


---

Por que usar pré-processamento?

Porque references.json.gz é pesado para carregar em runtime.

No build, eu converto:

JSON gzip -> binário float32 + labels

No runtime, o vector-engine só lê arquivos binários simples.


---

Por que usar float32?

O dataset do desafio trabalha com valores normalizados. Para esse tipo de comparação vetorial, float32 reduz memória e é suficiente para ordenar as distâncias dos vizinhos.

Além disso, usar float32 diminui o volume carregado em memória quando comparado com float64.


---

Por que não usar banco de dados?

Porque o desafio não pede persistência transacional.

O problema é puramente:

entrada JSON -> vetor -> busca KNN -> resposta JSON

Adicionar banco de dados aqui aumentaria complexidade e consumo de recurso sem resolver uma necessidade do enunciado.


---

Por que não adicionar heurísticas próprias?

Porque o resultado precisa seguir a proposta do desafio.

Eu não adiciono:

if MCC alto então fraude
if valor alto então fraude
if distância alta então recusar

A decisão final vem apenas dos 5 vizinhos mais próximos e do cálculo de fraud_score.


---

Por que não usar cache de resposta?

Porque o enunciado não exige e porque cada transação tem identificador e características próprias.

Um cache poderia até ajudar em cenários repetidos, mas também adicionaria uma regra operacional fora do núcleo do desafio. Nesta versão, preferi manter a solução limitada e previsível.


---

Limitações conhecidas

Esta versão usa busca exata por força bruta no vector-engine.

Isso tem uma vantagem:

resultado fiel ao KNN exato

Mas também tem um custo:

cada requisição compara contra todos os vetores de referência

Para o desafio real, o desempenho final precisa ser validado com o teste oficial. Eu não afirmo throughput ou latência sem executar a carga oficial no ambiente correto.


---

Comandos úteis

Subir:

docker compose up -d

Ver logs:

docker compose logs -f

Derrubar:

docker compose down

Rebuild completo:

docker compose down
docker build --target api -t nithack/rinha-2026-api:latest .
docker build --target vector-engine -t nithack/rinha-2026-vector-engine:latest .
docker compose up -d

Testar /ready:

curl -i http://localhost:9999/ready

Testar /fraud-score:

curl -s http://localhost:9999/fraud-score \
  -H 'Content-Type: application/json' \
  -d @payload.json


---

Publicação das imagens

Antes de submeter, eu preciso trocar o nome das imagens no docker-compose.yml para o registry correto.

Exemplo:

api1:
  image: meu-registry/rinha-2026-api:latest

api2:
  image: meu-registry/rinha-2026-api:latest

vector-engine:
  image: meu-registry/rinha-2026-vector-engine:latest

Build e push:

docker build --target api -t meu-registry/rinha-2026-api:latest .
docker build --target vector-engine -t meu-registry/rinha-2026-vector-engine:latest .

docker push meu-registry/rinha-2026-api:latest
docker push meu-registry/rinha-2026-vector-engine:latest


---

Resumo da solução

Eu mantive a solução focada no que o desafio pede:

2 instâncias de API
1 load balancer na porta 9999
1 serviço interno de busca vetorial
pré-processamento dos arquivos oficiais
vetorização com 14 dimensões
busca dos 5 vizinhos mais próximos
fraud_score = fraudes / 5
approved = fraud_score < 0.6

A implementação evita dependências desnecessárias e mantém o caminho crítico pequeno:

Nginx -> API -> Vector Engine -> API -> resposta

A ideia é competir com uma solução simples de entender, fácil de revisar e diretamente conectada ao enunciado.
