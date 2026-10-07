# Fit

https://github.com/user-attachments/assets/440e14dd-066b-496a-ba32-7b914aba34f2

Fit é um sistema local de controle de versão distribuído.

Ao final dessa especificação, encontra-se um glossário para os termos-chave.

Para controlar as expectativas, vamos considerar que essa especificação se aplica para uma primeira versão do projeto, chamada ao longo desse documento de v1.

As definições a seguir fazem comparação de conceitos entre Git e Fit, como blobs, commits, etc.

## Integração com Docker

Requer Docker com Buildx/Bake; no Windows, use Docker Desktop com containers Linux.

Não é necessário instalar Go.

Na raiz do projeto:

```sh
cp .env.example .env
# Edite .env com uma senha aleatória antes de iniciar o broker.
docker buildx bake
docker compose up -d --build --wait rabbitmq
```

> Para compilar só uma versão: `docker buildx bake linux-amd64` ou
`docker buildx bake windows-amd64`.

Execute o binário no diretório dos arquivos que deseja versionar.

No PowerShell, partindo da raiz do projeto:

```powershell
$env:FIT_BROKER_URL = "amqp://fit:SENHA@localhost:5672/fit"
$fit = (Resolve-Path .\dist\windows-amd64\fit.exe).Path
New-Item -ItemType Directory -Force "$HOME\fit-demo" | Out-Null
Set-Location "$HOME\fit-demo"
& $fit init
& $fit serve
```

No Linux:

```bash
export FIT_BROKER_URL="amqp://fit:SENHA@localhost:5672/fit"
fit=$(realpath dist/linux-amd64/fit)
mkdir -p ~/fit-demo
cd ~/fit-demo
$fit init
$fit serve
```

## Objetivo

- commits e histórico em DAG (vide [`## Modelo de histórico`](#modelo-de-histórico));
- replicação entre peers;
- trabalho offline;
- consistência eventual;
- divergência, merge e conflitos;
- tolerância a falhas de peers e do transporte.

O ambiente da v1 assume uma máquina, uma instância RabbitMQ, vários processos Fit e diretórios independentes.

O foco é demonstrar comunicação assíncrona, descoberta e reconciliação via mensageria, não atingir escala de produção. Exchanges por `repositoryId` segmentam o tráfego; o fan-out ainda entrega cada pedido a todos os peers daquele repositório. Cluster RabbitMQ, alta disponibilidade e particionamento adicional são possibilidades futuras e não foram implementados.

## Segurança do RabbitMQ

O cliente exige `FIT_BROKER_URL`, sem fallback para `guest:guest`. O Compose exige usuário e senha em `.env` (não versionado) e usa o vhost `fit`. Use senha aleatória e codifique caracteres reservados da URL; em produção, crie credenciais individuais por peer.

O script de inicialização remove tags administrativas do usuário e restringe suas permissões a recursos `fit.*`, filas temporárias `amq.gen-*` e publicação na exchange padrão `amq.default` para request/reply. O usuário não recebe acesso a outros vhosts. As portas estão limitadas a localhost; o usuário da aplicação não acessa o painel de administração. As credenciais iniciais só são criadas em um volume novo; volumes existentes precisam de migração explícita com `rabbitmqctl`.

A demonstração básica usa AMQP local. Para criptografar o transporte, coloque `ca.pem`, `server.pem` e `server-key.pem` em `docker/certs/` (não versionado), com chave legível pelo usuário `rabbitmq` do container. O certificado do servidor precisa de SAN `DNS:localhost` e assinatura pela CA; não desative a validação de certificados.

```sh
docker compose -f compose.yaml -f compose.tls.yaml up -d --build --wait
export FIT_BROKER_URL="amqps://fit:SENHA@localhost:5671/fit"
export FIT_BROKER_CA_FILE="/caminho/absoluto/ca.pem"
```

A configuração TLS desativa o listener AMQP sem criptografia e permite TLS 1.2/1.3. O cliente valida a CA e o hostname; a autenticação do peer continua por usuário/senha, sem exigir certificado de cliente. Isso protege o trecho peer–broker; criptografia ponta a ponta entre peers continua fora do escopo.

## Conceitos

### Peer

Um peer é um diretório de trabalho que contém um diretório `.fit` e um processo Fit associado.

A presença de `.fit` identifica o peer.

Cada peer possui:

- `peerId` próprio e persistente;
- o mesmo `repositoryId` dos demais peers daquele repositório;
- armazenamento local de commits;
- `HEAD` próprio;
- working tree independente.

Ao reiniciar, o peer mantém sua identidade e recupera seu estado a partir de `.fit`.


Ou seja, o estado autoritativo vive nos peers. RabbitMQ transporta anúncios, pedidos e respostas, mas não é o banco de dados do repositório.

### Repositório

`fit init` cria o primeiro peer e um novo repositório:

- `repositoryId`: UUID, identidade real do repositório;
- `repositoryName`: nome humano, não necessariamente único;
- `peerId`: UUID do peer criador.

Todos os peers são equivalentes. Não existe servidor central, peer líder ou peer proprietário.

### Tipo de peer

A v1 implementa apenas o developer peer: recebe anúncios enquanto conectado, baixa conteúdo com `sync` e atualiza HEAD/working tree por ação explícita do usuário.

## Arquitetura

```mermaid
flowchart TB
    CLI["CLI: add, commit, sync, merge"]
    Repo["Repository: index, commits, HEAD"]
    Sync["Sync Service: discovery e reconciliação"]
    MQ["RabbitMQ Transport"]

    CLI --> Repo
    Repo <--> Sync
    Sync <--> MQ
```

O transporte deve ser isolado por uma interface, permitindo demonstrar o `swap` da implementação da camada MOM - sem `vendor lock-in` -, mesmo que `RabbitMQTransport` seja a única implementação da v1.

## Modelo de histórico

- Não existem branches na v1.
- Cada peer possui um `HEAD` local.
- Heads remotos conhecidos são armazenados por `peerId`.
- O histórico é um DAG.
- Commit comum possui zero ou um pai; merge commit possui dois pais.
- Merge com três ou mais heads é feito de forma binária, um head por vez.
- Timestamps são informativos; causalidade e ancestralidade são determinadas apenas pelo DAG.

```mermaid
gitGraph
    commit id: "C1"
    branch peer-b
    checkout main
    commit id: "C2A"
    checkout peer-b
    commit id: "C2B"
    checkout main
    merge peer-b id: "M"
```

## Uso do MOM

A introdução de uma camada responsável pelo transporte de mensagens, ao invés de comunicação direta entre peers, permite:

## Vantagens

- desacoplamento entre peers.
- descoberta de peers simplificada.
- fan-out natural.
- delegação parcial de confiabilidade ao broker.

## Desvantagens

- ponto central de falha (vai em desencontro com a descentralização natural do P2P).
- latência adicional devido ao caminho indireto.
- complexidade operacional adicional.
- limitações relativas ao broker (escalabilidade, tamanho máximo de mensagem, etc.).

Observação: a v1 considera a transmissão de blobs através do broker.

Isso não é ideal, mas simplifica a demonstração. Em versões futuras, o transporte de blobs pode ser feito entre peers diretamente, não invalidando a presença do broker para outras operações.

## Topologia RabbitMQ

Na v1, RabbitMQ atua apenas como camada de transporte entre os peers. O estado do repositório continua armazenado localmente em cada `.fit`.

A topologia utiliza:

* uma exchange global `fit.discovery`, para descoberta de repositórios;
* uma exchange `fit.repo.<repositoryId>` por repositório;
* uma fila efêmera por processo Fit conectado;
* filas temporárias de resposta (`reply-to`) para operações como `clone` e `sync`.

### Dead letter queue e retry

As filas de assinatura usam `x-dead-letter-exchange=fit.dlx`; essa exchange `fanout` durável encaminha rejeições (`Nack(false)`) para a fila compartilhada `fit.dlq`, também durável. Mensagens inválidas ou falhas de processamento ficam disponíveis para inspeção manual, sem requeue infinito. A DLQ retém mensagens por até 24 horas e no máximo 10.000 entradas; não há replay automático. A durabilidade da fila não garante persistência de mensagens transitórias após reinício do broker.

O consumidor usa ACK manual e prefetch 16. O retry ocorre no solicitante, com até três publicações e backoff de 1 s/2 s, respeitando o contexto; não é um ciclo de redelivery da DLQ. As respostas e anúncios permanecem efêmeros, recuperáveis por novos pedidos.

### Visão geral

```mermaid
flowchart LR
    A["Peer A"] --> X{{"fit.repo.<repositoryId>"}}
    B["Peer B"] --> X

    X --> QA[("fila A")]
    X --> QB[("fila B")]

    QA --> A
    QB --> B
```

A **exchange** recebe mensagens e as distribui para as filas associadas. Cada processo Fit consome sua própria fila temporária.

### Descoberta

`fit.discovery` permite localizar peers que possuam determinado repositório.

```mermaid
flowchart LR
    C["fit clone / fit repos"] -->|"repository.discover"| D{{"fit.discovery"}}
    D --> Q[("filas dos peers")]
    Q --> P["Peers"]
    P -->|"repository.offer"| R[("reply-to")]
    R --> C
```

A descoberta é ativa: RabbitMQ não mantém um catálogo permanente dos repositórios.

### Comunicação de um repositório

Mensagens específicas de um repositório passam pela exchange:

```text
fit.repo.<repositoryId>
```

Por exemplo:

```mermaid
flowchart LR
    A["Peer A"] -->|"head.announce(C5)"| X{{"fit.repo.<repositoryId>"}}

    X --> QB[("fila Peer B")]
    X --> QR[("fila Peer C")]

    QB --> B["Peer B"]
    QR --> R["Peer C"]
```

Esse mesmo mecanismo é usado para mensagens como `head.request`, `commit.request` e `blob.request`.

### Request e response

Requests podem ser recebidos por vários peers. As respostas são enviadas diretamente para a fila indicada em `reply-to`.

```mermaid
flowchart LR
    B["Peer B"] -->|"commit.request(C5)"| X{{"fit.repo.<repositoryId>"}}
    X --> A["Peer A"]
    A -->|"commit.response(C5)"| R[("reply-to B")]
    R --> B
```

O `correlationId` associa cada resposta ao request correspondente.

Quando vários peers respondem, a primeira resposta válida é aceita e as demais são ignoradas.

### Filas efêmeras e recuperação

As filas são temporárias. Se um processo desconectar, mensagens pendentes podem ser perdidas.

Isso não compromete o repositório, porque o estado persistente está nos peers.

```mermaid
flowchart LR
    A["Peer A<br/>.fit"] <-->|"mensagens"| MQ["RabbitMQ"]
    B["Peer B<br/>.fit"] <-->|"mensagens"| MQ
```

Ao reconectar ou executar `sync`, o peer publica `head.request`, permitindo reconstruir o conhecimento sobre os heads atuais.

Portanto, RabbitMQ transporta anúncios, requests e responses, enquanto commits, blobs, `HEAD` e refs permanecem armazenados nos peers.

### Uso do `reply-to`

A decisão tomada na v1 é uso de filas efêmeras para aceitar respostas de requests.

Ou seja, requests como `fit clone` ou `fit sync` criam uma fila temporária, que é destruída ao final da operação. O peer que responde envia a resposta para essa fila.

Já o uso de `fit serve` mantém uma fila temporária para receber anúncios e requests de outros peers, enquanto o processo estiver ativo.

Então, filas efêmeras são bastante importantes para a v1, porque permitem que requests e responses sejam tratados de forma assíncrona, sem exigir que o peer esteja sempre online.

## Modelo de commit

Cada commit representa um `snapshot` lógico completo: metadados, pais e um manifesto `path -> blobHash`. O conteúdo dos arquivos fica em blobs separados, identificados pelo SHA-256 dos seus conteúdos. Não existem objetos `tree` na v1.

Arquivos inalterados reutilizam o mesmo blob entre commits. Alterar um arquivo entre 100 gera um novo manifesto e, se o conteúdo ainda não existir, um único blob novo. A rede transfere somente commits e blobs ausentes no destinatário, não o repositório inteiro a cada commit.

```json
{
  "id": "sha256:...",
  "repositoryId": "uuid",
  "parents": ["sha256:..."],
  "author": {
    "peerId": "uuid"
  },
  "timestamp": "2026-09-17T12:00:00Z",
  "message": "adiciona configuração",
  "files": {
    "README.md": "sha256:hash-do-readme",
    "config.txt": "sha256:hash-da-config"
  }
}
```

Regras:

- o ID é SHA-256 da representação canônica do commit sem o campo `id`;
- a serialização do commit deve ser determinística: JSON em UTF-8, chaves ordenadas recursivamente e sem espaços opcionais;
- o hash do blob usa os bytes originais, sem normalizar encoding ou finais de linha;
- o manifesto lista todos os arquivos rastreados; ausência de um path representa deleção, sem apagar blobs históricos;
- paths devem ser relativos, normalizados e não podem escapar do working tree;
- hashes são recalculados antes de persistir commits ou blobs recebidos;
- commits e respostas duplicadas são idempotentes;
- arquivos binários podem ser armazenados, mas não recebem merge textual.

O commit não contém conteúdo de arquivo, nem em texto nem em base64. Checkout e merge consultam o manifesto e leem os blobs referenciados. Um commit com hash válido pode estar armazenado enquanto seus blobs ainda estão pendentes; nesse caso, ele não está pronto para aplicação.

Para merge, um blob é considerado texto quando contém UTF-8 válido; caso contrário, é tratado como binário. O hash continua sendo calculado sobre os bytes originais.

## Estrutura local

```text
.fit/
├── config.json
├── HEAD
├── index.json
├── commits/
│   └── <commit-id>.json
├── blobs/
│   └── <blob-hash>
├── refs/
│   └── peers/
│       └── <peer-id>.json
├── SYNC_HEADS.json
├── MERGE_HEAD
└── MERGE_BASE
```

- `config.json`: `formatVersion`, `repositoryId`, `peerId` e `repositoryName`. O broker é configurado por variáveis de ambiente, fora do estado versionado do peer.
- `HEAD`: commit aplicado localmente.
- `index.json`: paths preparados com hashes de blobs e deleções registradas por `fit add` e `fit rm`.
- `commits/`: commits validados e imutáveis.
- `blobs/`: bytes imutáveis dos arquivos, deduplicados por hash. Os nomes físicos usam o digest hexadecimal, sem o prefixo `sha256:`.
- `refs/peers/`: último head conhecido de cada peer.
- refs remotas não expiram automaticamente na v1; disponibilidade é determinada por discovery ou respostas recentes.
- `SYNC_HEADS.json`: heads obtidos na última sincronização e indicação de download completo ou pendente.
- `MERGE_HEAD` e `MERGE_BASE`: merge em andamento e sua base.

Escritas em `.fit` devem usar arquivo temporário e rename atômico sempre que possível.

`config.json` deve conter um `formatVersion`; mensagens devem conter um `protocolVersion`. A v1 rejeita versões incompatíveis com erro explícito.

O lock local deve serializar escritas concorrentes entre CLI e `serve` e ser liberado automaticamente quando o processo termina.

## Comandos

### Criação e descoberta

```bash
fit init
fit repos
fit clone <repository-id|nome> <diretório>
fit serve
```

- `init`: cria `.fit` e gera os IDs; o nome é inferido do diretório atual. A publicação na rede ocorre quando `serve` é iniciado.
- `repos`: faz descoberta ativa e lista ofertas de peers vivos.
- `clone`: descobre o repositório, escolhe uma oferta válida, baixa os commits alcançáveis pelo head oferecido e todos os blobs referenciados por eles, e materializa esse head. Blobs repetidos são baixados uma única vez.
- `serve`: mantém este peer conectado e processando mensagens, em primeiro plano, até `Ctrl+C`.

### Processo de rede

`init` cria o peer no disco; `serve` o coloca online. São operações distintas: `.fit` continua existindo quando o processo termina.

Execute `fit serve` em um terminal dentro do diretório do peer. Em outro terminal, no mesmo diretório, execute `add`, `commit`, `sync` e `merge`. O serviço anuncia mudanças de HEAD, recebe anúncios e responde a pedidos de commits e blobs.

Não é necessário executar `serve` para trabalhar localmente. Sem ele, o peer não fica disponível continuamente para outros peers. `repos`, `clone` e `sync` podem abrir uma conexão temporária durante o comando, sem exigir um serviço já iniciado. Cada peer permite apenas um `serve` ativo; CLI e serviço serializam escritas em `.fit` por lock local.

Estar ouvindo não significa aplicar mudanças: no developer peer, `serve` apenas registra anúncios e atende pedidos. A atualização dos arquivos continua explícita.

Se um nome corresponder a mais de um `repositoryId`, o clone exige o ID explícito. Se vários peers oferecerem o mesmo repositório, a primeira oferta válida pode ser usada; outra oferta pode assumir em caso de timeout ou conteúdo inválido.

### Trabalho local

```bash
fit add <path...>
fit rm <path...>
fit commit -m <mensagem>
fit status
fit checkout <commit-id>
```

- `add`: grava o conteúdo atual como blob, se ainda não existir, e registra seu hash no index. Edições posteriores não alteram o conteúdo preparado.
- `rm`: remove o arquivo do working tree e registra sua ausência no index.
- `commit`: cria o manifesto completo combinando o manifesto de `HEAD` e as alterações do index. Verifica a presença dos blobs referenciados, persiste o commit, move `HEAD`. Quando `serve` está executando, ele detecta a mudança e anuncia o novo head; `commit` não publica diretamente.
- `checkout`: substitui o working tree pelo snapshot escolhido; exige working tree limpo.

Com merge conflitante em andamento, `commit` só é permitido depois que todos os conflitos forem resolvidos e adicionados ao index. O commit resultante terá dois pais.

### Sincronização

```bash
fit sync
fit merge <commit-id>
fit merge --abort
```

- `sync`: consulta os heads dos peers disponíveis e baixa conteúdo faltante. Mostra os IDs disponíveis para integração. Não altera `HEAD`, index nem working tree do developer peer.
- `merge <commit-id>`: integra um commit já baixado. Faz fast-forward quando possível; caso contrário, executa 3-way merge e cria um merge commit ou registra conflitos. Exige index e working tree limpos e nenhum merge em andamento. Não acessa a rede.
- `merge --abort`: cancela um merge em andamento e restaura `HEAD`, index e working tree ao estado anterior ao merge.

`fetch` e `pull` não existem na API v1. O fluxo é `fit sync`, seguido de `fit merge <commit-id>` quando o usuário quiser aplicar uma atualização. Com vários heads divergentes, a escolha do commit é explícita.

Receber um anúncio atualiza conhecimento; `sync` baixa conteúdo; `merge` integra e aplica; `checkout` materializa um commit escolhido sem integrar históricos.

### Quando a sincronização está completa

Para cada head coletado, `sync` percorre todos os pais até a raiz e solicita os commits ausentes. Em seguida, solicita apenas os blobs faltantes nos manifestos desse histórico, deduplicando pedidos por hash. Um commit já presente não dispensa verificar seus pais e blobs.

Um head só é marcado como completo quando todos os commits alcançáveis e seus blobs estiverem validados e persistidos. Timeout deixa o head pendente em `SYNC_HEADS.json`; nova execução retoma o que falta. Heads anunciados depois da coleta ficam para a próxima sincronização.

`merge` e `checkout` verificam as dependências antes de alterar arquivos ou `HEAD`. Se houver conteúdo faltante, interrompem sem aplicar parcialmente e indicam a necessidade de sincronização.

## Descoberta e clone

RabbitMQ possui:

- uma exchange global de descoberta `fit.discovery`;
- uma exchange por repositório `fit.repo.<repositoryId>`;
- filas temporárias dos peers e das operações de curta duração.

```mermaid
sequenceDiagram
    participant B as Novo peer B
    participant MQ as RabbitMQ
    participant A as Peer A

    B->>MQ: repository.discover(nome ou id)
    MQ->>A: repository.discover
    A->>MQ: repository.offer(repoId, peerId, head)
    MQ->>B: repository.offer
    B->>MQ: commit.request(head)
    MQ->>A: commit.request
    A->>MQ: commit.response(commit)
    MQ->>B: commit.response
    Note over B: repete para pais faltantes
    B->>MQ: blob.request(hash ausente)
    MQ->>A: blob.request
    A->>MQ: blob.response(bytes)
    MQ->>B: blob.response
    Note over B: repete para blobs faltantes
    B->>B: valida, persiste e checkout
```

A descoberta é ativa: `repository.discover` recebe `repository.offer` dos peers conectados. Não há registro permanente de presença.

## Protocolo de mensagens

Envelope JSON para mensagens de controle e respostas de commit:

```json
{
  "protocolVersion": 1,
  "messageId": "uuid",
  "type": "head.announce",
  "repositoryId": "uuid",
  "senderPeerId": "uuid",
  "correlationId": "uuid-opcional",
  "sentAt": "2026-09-17T12:00:00Z",
  "payload": {}
}
```

Mensagens mínimas:

| Mensagem | Finalidade |
| --- | --- |
| `repository.discover` | Procurar repositório por nome ou ID |
| `repository.offer` | Oferecer repo, peer e head para clone |
| `head.announce` | Informar o head atual do peer |
| `head.request` | Solicitar que peers vivos anunciem seus heads atuais |
| `commit.request` | Solicitar um commit pelo ID |
| `commit.response` | Entregar metadados e manifesto do commit, sem conteúdo dos arquivos |
| `blob.request` | Solicitar um blob pelo hash |
| `blob.response` | Entregar os bytes de um blob |

`blob.response` usa corpo binário (`application/octet-stream`), sem base64. Os campos de identificação do envelope e `blobHash` seguem nas headers AMQP; `correlationId` associa a resposta ao pedido. Cada resposta carrega um blob inteiro. Fragmentação e streaming ficam fora da v1; o limite de mensagem é o configurado no broker.

Regras do protocolo:

- mensagens podem ser perdidas, duplicadas ou chegar fora de ordem;
- qualquer peer do repositório que possua o commit ou blob solicitado pode responder;
- a primeira resposta válida é aceita; as demais são ignoradas;
- ao receber um commit cujo pai esteja ausente, o peer solicita o pai;
- manifestos determinam os blobs necessários; blobs locais válidos não são solicitados novamente;
- requests são publicados no máximo três vezes: imediatamente, após 1 s e após mais 2 s; a última janela de resposta dura 4 s. O contexto do comando pode encerrar a coleta antes disso; o mesmo `reply-to` e `correlationId` são reutilizados. As operações atendidas são consultas idempotentes;
- requests que esperam resposta informam uma fila de retorno (`reply-to`);
- retries são por objeto solicitado; objetos já validados não são baixados novamente;
- o destinatário valida `protocolVersion`, `repositoryId`, hash, estrutura e paths;
- um anúncio nunca altera automaticamente o working tree de developer peers.
- ao iniciar `sync` ou reconectar, publicar `head.request`; cada peer responde com `head.announce`. A coleta tem timeout: silêncio não significa repositório atualizado. O comando informa fontes indisponíveis e downloads pendentes.

## Reconciliação

Ao comparar o `HEAD` local com um head remoto conhecido:

| Relação | Estado | Ação de merge |
| --- | --- | --- |
| Iguais | `CLEAN` | Nenhuma |
| Local é ancestral do remoto | `BEHIND` | Fast-forward |
| Remoto é ancestral do local | `AHEAD` | Nenhuma |
| Nenhum é ancestral do outro | `DIVERGED` | 3-way merge |

```mermaid
flowchart TD
    A["Comparar HEAD local e remoto"] --> B{"Relação no DAG"}
    B -->|iguais| C["CLEAN"]
    B -->|local é ancestral| D["BEHIND: fast-forward"]
    B -->|remoto é ancestral| E["AHEAD"]
    B -->|divergentes| F["3-way merge"]
    F -->|sem conflito| G["Merge commit"]
    F -->|com conflito| H["CONFLICTED"]
```

## Merge e conflitos

O merge usa:

- `base`: ancestral comum mais próximo;
- `ours`: `HEAD` local;
- `theirs`: commit informado pelo usuário.

Na v1, se houver mais de uma merge base possível, o comando falha com diagnóstico explícito.

Conflitos suportados:

- `modify/modify`;
- `modify/delete`;
- `add/add` com conteúdos diferentes;
- alterações concorrentes em arquivo binário.

Arquivos de texto usam 3-way merge por linha. Em conflito `modify/modify`, o working tree recebe marcadores:

```text
<<<<<<< ours
valor = A
=======
valor = B
>>>>>>> theirs
```

`modify/delete`, `add/add` e conflitos binários são registrados no index e apresentados por `fit status`; conteúdo não é escolhido silenciosamente.

Durante um conflito:

1. `MERGE_HEAD` e `MERGE_BASE` são persistidos antes de alterar o working tree, junto do estado necessário para `merge --abort`.
2. O usuário resolve os arquivos.
3. Executa `fit add` ou `fit rm` em cada conflito.
4. Executa `fit commit -m "resolve conflito"`.
5. O novo commit possui `[ours, theirs]` como pais.
6. O peer anuncia o merge commit como qualquer outro novo head.

Peers que já estejam em qualquer um dos pais podem integrar o merge commit por fast-forward.

## Estados do repositório

| Estado | Significado |
| --- | --- |
| `CLEAN` | Working tree igual ao `HEAD`, sem atualização pendente |
| `DIRTY` | Working tree ou index difere do `HEAD` |
| `BEHIND` | Há head remoto descendente do local |
| `AHEAD` | Head local descende do remoto |
| `DIVERGED` | Heads local e remoto formam linhas concorrentes |
| `MERGING` | Existe merge em andamento |
| `CONFLICTED` | Merge possui conflitos não resolvidos |
| `TRANSPORT_OFFLINE` | Broker indisponível; operações locais continuam válidas |

Estados podem coexistir quando não forem contraditórios, por exemplo `DIRTY + TRANSPORT_OFFLINE`.

## Política dos peers

### Developer peer

Ao receber `head.announce`:

1. registra o head remoto;
2. não baixa commits automaticamente;
3. não altera `HEAD` nem working tree;
4. aguarda `sync` para baixar e `merge` para aplicar.

## Falhas e recuperação

### RabbitMQ indisponível

- `add`, `rm`, `commit`, `status` e operações locais continuam funcionando.
- Discovery, anúncios e transferências ficam indisponíveis.
- Commits criados offline permanecem no `.fit`.
- Ao reconectar, o peer anuncia `HEAD` atual e responde à descoberta.
- Não é necessário reproduzir anúncios perdidos.

### Reinício do RabbitMQ

- peers recriam exchanges/filas temporárias e anunciam novamente `HEAD` e respondem à descoberta;
- a recuperação não depende das mensagens existentes antes da queda.

### Peer indisponível

- Os demais peers continuam operando.
- Ao reiniciar, o peer relê `.fit`, mantém `peerId` e anuncia seu estado atual.
- Trabalho feito offline é reconciliado pelo DAG após reconexão.

### Queda durante sync

- Cada commit ou blob validado é persistido de forma atômica.
- Na próxima tentativa, objetos válidos existentes são reutilizados e apenas os faltantes são solicitados, inclusive blobs de commits já recebidos.
- Um download parcial nunca é aceito como objeto válido. Ter o manifesto não significa ter seu conteúdo completo.

### Queda durante merge

- `MERGE_HEAD` e `MERGE_BASE` permitem retomar o merge.
- Após reiniciar, `fit status` informa `MERGING` ou `CONFLICTED`.
- O peer não inicia outro merge até concluir ou abortar o atual.

### Mensagens duplicadas ou fora de ordem

- Operações são idempotentes por `commitId`, `blobHash` e estado atual.
- Commits podem chegar antes dos pais; pais ausentes são solicitados.
- Heads concorrentes nunca sobrescrevem uns aos outros silenciosamente.

### Mensagens perdidas

- anúncios perdidos são recuperados por discovery ou `head.request`;
- requests e responses perdidos são tratados por timeout e retry limitado.

### Clone interrompido

- O diretório incompleto é marcado como clone em andamento.
- Uma nova execução retoma os commits e blobs faltantes ou remove apenas o estado parcial do clone.
- Se nenhum peer responder, o comando termina com timeout sem criar um peer válido.

### Conteúdo inválido

- Commit com hash incorreto, `repositoryId` divergente ou path inseguro é rejeitado.
- Blob cujos bytes não correspondam ao hash solicitado é rejeitado e não libera a aplicação do commit.
- O peer pode tentar outra fonte.
- O erro é registrado e exibido sem corromper o repositório local.

## Fluxo principal da demonstração

```mermaid
sequenceDiagram
    participant A as Developer A
    participant MQ as RabbitMQ
    participant B as Developer B

    A->>A: commit C2A offline
    B->>B: commit C2B offline
    A->>MQ: head.announce(C2A)
    B->>MQ: head.announce(C2B)
    B->>B: Executar sync e merge C2A
    B->>B: Resolver conflito e criar commit M
    B->>MQ: head.announce(M)
    MQ->>A: head M disponível
    A->>A: Executar sync e merge M por fast-forward
```

Roteiro mínimo:

1. Criar repo com `fit init`.
2. Executar `fit serve` em A, descobrir com `fit repos` e criar outros peers com `fit clone`. Iniciar `fit serve` em cada peer, em terminais separados.
3. Demonstrar propagação e fast-forward.
4. Parar processos ou RabbitMQ.
5. Criar commits divergentes em A e B.
6. Restaurar o mesh e executar `fit sync`.
7. Exibir `DIVERGED`, executar merge e resolver conflito.
8. Anunciar o merge commit e mostrar fast-forward nos outros peers.

## Funcionalidades não implementadas

- **Replica peer:** foi descartado da entrega por falta de tempo. Download e fast-forward automáticos não foram implementados; todos os peers seguem a política developer.
- **`repository.announce` e `peer.announce`:** durante a implementação, o uso dessas mensagens deixou de fazer sentido. A descoberta ativa com `repository.discover`/`repository.offer` e os anúncios de `head.announce` atendem ao fluxo utilizado.
- **`fit log`:** ficou de fora por falta de tempo; não é um comando disponível na CLI.
- **Broker em `config.json`:** não foi incluído porque não fez sentido persistir configuração de ambiente junto à identidade do peer. A conexão usa `FIT_BROKER_URL` e, opcionalmente, `FIT_BROKER_CA_FILE`.

## Fora do escopo da v1

- validação da identidade de peers não confiáveis além das credenciais do broker;
- criptografia ponta a ponta;
- branches, tags e rebase;
- garbage collection e compactação;
- objetos tree, deltas, packfiles e fragmentação de blobs;
- múltiplas máquinas e descoberta fora de localhost;
- consenso, líder global ou ordem total de commits;
- merge automático de arquivos binários;
- resolução automática de múltiplas merge bases;
- compatibilidade entre versões diferentes do formato ou protocolo.

## Critérios de aceite

A v1 está completa quando:

- peers criam commits sem RabbitMQ;
- um repositório pode ser descoberto e clonado pelo mesh;
- commits contêm manifestos, nunca o conteúdo embutido dos arquivos;
- alterar um arquivo reutiliza os blobs inalterados, sem retransmiti-los a peers que já os possuem;
- clone e sync verificam commits, ancestrais e blobs antes de declarar um head completo;
- blobs corrompidos ou ausentes impedem aplicação sem alterar parcialmente o working tree;
- developer peers separam anúncio, download via `sync` e aplicação via `merge`;
- commits duplicados e fora de ordem não corrompem estado;
- divergência é detectada pelo DAG;
- merge textual cria commit com dois pais;
- conflitos sobrevivem a restart;
- merge em andamento pode ser abortado;
- perda de mensagens não impede convergência posterior;
- versões incompatíveis são rejeitadas explicitamente;
- após falhas do broker ou de peers, a troca de heads permite convergência sem replay completo de eventos.

## Glossário

- DAG: Directed Acyclic Graph, grafo acíclico direcionado. No Fit, o histórico de commits forma um DAG, onde cada commit aponta para seus pais, e não há ciclos.

