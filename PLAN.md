# FIT — Plano Completo de Implementação em Go, guiado por TDD

Este plano parte de uma premissa simples:

> A implementação não deve ser desenhada inteira antes do código existir. O sistema deve ser descoberto por meio de testes pequenos, incrementais e comportamentais.

A sequência de trabalho para cada passo é:

```text
RED
escrever o menor teste que expressa o próximo comportamento

GREEN
fazer passar da forma mais simples possível

REFACTOR
reorganizar somente código que já existe

COMMIT
commitar teste + implementação

próximo comportamento
```

Uma abstração nova só entra quando:

- dois ou mais usos reais revelam o mesmo conceito;
- a duplicação começa a prejudicar mudanças;
- o acoplamento torna testes difíceis;
- a implementação atual deixa uma mudança insegura.

Não criar interfaces, services, repositories, adapters ou camadas porque “vamos precisar depois”.

---

# 1. Blob identity

## Primeiro teste

```go
func TestBlobIDUsesSHA256(t *testing.T) {
    got := BlobID([]byte("hello"))

    want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"

    if got != want {
        t.Fatalf("got %q, want %q", got, want)
    }
}
```

## Edge cases

- `[]byte{}` produz o SHA-256 correto do conteúdo vazio;
- `"a"` e `"A"` produzem hashes diferentes;
- `"\n"` e `"\r\n"` produzem hashes diferentes;
- UTF-8 é tratado como bytes, sem normalização;
- bytes inválidos como UTF-8 continuam válidos;
- arquivo grande produz o mesmo hash independentemente da forma de leitura;
- repetir `BlobID` com os mesmos bytes é determinístico.

## Critério para avançar

A identidade de conteúdo deve ser completamente determinística.

---

# 2. Persistência de blobs

## Testes mínimos

```text
PutBlob armazena bytes pelo hash
GetBlob recupera bytes pelo hash
```

## Edge cases

- gravar o mesmo blob duas vezes não falha;
- segundo `PutBlob` não altera o conteúdo existente;
- hash inexistente retorna erro distinguível;
- conteúdo físico incorreto para aquele hash é rejeitado;
- arquivo vazio funciona;
- blob binário com `0x00` funciona;
- conteúdo arbitrário não é tratado como string;
- diretório `blobs/` inexistente é criado;
- erro de filesystem é propagado;
- arquivo temporário abandonado não é tratado como blob;
- `<hash>.tmp` não é encontrado por `GetBlob`;
- digest físico usa hexadecimal simples.

## Refactor possível

```go
type BlobID string
```

Não criar `BlobStore` interface ainda.

---

# 3. Paths seguros

Antes de `add`, validar paths.

## Happy path

```text
foo.txt
src/main.go
a/b/c.txt
```

## Edge cases

```text
../foo
../../foo
/absolute/foo
./foo
foo/../bar
foo//bar
.
..
```

Também testar:

- acesso a `.fit`;
- symlink que resolve para fora do working tree;
- diretório em vez de arquivo;
- arquivo inexistente;
- mesmo arquivo por paths equivalentes;
- path vazio;
- Unicode;
- espaços;
- arquivos ocultos.

## Propriedade importante

```text
normalize(normalize(path)) == normalize(path)
```

---

# 4. `add`

## Primeiro teste semântico

```text
arquivo contém A
add arquivo
arquivo muda para B
staged content continua A
```

## Edge cases

- arquivo novo;
- arquivo já tracked;
- `add` duas vezes sem mudança;
- `add` novamente depois de mudar;
- dois arquivos;
- `add` parcial preserva outras entries;
- mesmo conteúdo em dois paths reutiliza blob;
- arquivo vazio;
- binário;
- arquivo grande;
- arquivo inexistente;
- diretório;
- `add` durante merge resolve path;
- `add` não altera `HEAD`;
- `add` não altera commits antigos.

## Caso concorrente importante

```text
leitura começa
arquivo muda durante leitura
```

Definir a semântica explicitamente.

---

# 5. Index

## Testes

```text
index vazio pode ser lido
uma entrada pode ser persistida
duas entradas sobrevivem a reload
```

## Edge cases

- index inexistente;
- JSON inválido;
- versão desconhecida;
- entry com `blob` e `delete`;
- entry sem operação válida;
- hash malformado;
- path inseguro;
- duplicidade lógica após normalização;
- atualizar uma entry preserva outras;
- write interrompido não produz JSON parcialmente válido;
- restart preserva staged state.

## Invariante

```text
index nunca contém o conteúdo dos arquivos
```

O index contém intenção staged e referências para blobs.

---

# 6. `rm`

## Happy path

```text
tracked file
fit rm
arquivo desaparece
deleção fica staged
```

## Edge cases

- remover tracked;
- remover já ausente;
- remover untracked;
- remover duas vezes;
- `rm` depois de `add`;
- `add` depois de `rm`;
- remover arquivo modificado e não staged;
- path inseguro;
- blobs históricos continuam existindo;
- commits antigos não mudam;
- restart preserva deleção staged.

---

# 7. Commit canônico

## Primeiro teste

```text
mesmo commit lógico => mesmo CommitID
```

## Edge cases

- ordem diferente das keys de `files` não muda ID;
- ordem diferente de campos internos não muda ID;
- espaços JSON não influenciam;
- `id` é excluído do hash;
- mensagem diferente muda ID;
- parent diferente muda ID;
- blob diferente muda ID;
- `repositoryId` diferente muda ID;
- autor diferente muda ID;
- timestamp diferente muda ID;
- encode/decode preserva o ID.

---

# 8. Persistência de commit

## Happy path

```text
PutCommit
GetCommit
```

## Edge cases

- commit duplicado é idempotente;
- arquivo com nome errado é rejeitado;
- JSON não bate com ID;
- commit pode ser persistido com blob ausente;
- commit pode ser persistido com parent ausente;
- path inseguro é rejeitado;
- `repositoryId` remoto incorreto é rejeitado;
- mais de dois parents;
- dois parents iguais;
- ciclo artificial no DAG;
- JSON truncado;
- campo obrigatório ausente;
- hash malformado;
- duplicate JSON keys.

---

# 9. Primeiro commit e HEAD

## Testes

- primeiro commit tem zero pais;
- `HEAD` aponta para ele;
- index é limpo;
- manifesto é completo;
- staged deletion não entra no manifesto;
- working tree não é relido durante commit;
- commit usa staged blob;
- commit sem mudanças tem semântica definida;
- blob staged ausente causa falha;
- blob staged corrompido causa falha;
- falha ao persistir commit não move HEAD;
- falha ao mover HEAD não limpa index;
- restart recupera HEAD.

## Ordem segura esperada

```text
validar
→ persistir commit
→ mover HEAD
→ limpar index
```

---

# 10. Segundo commit

## Cenário

```text
A:
foo → X
bar → Y

stage foo → X2

B:
foo → X2
bar → Y
```

## Testes

- arquivo não staged vem do manifesto anterior;
- arquivo modificado mas não staged mantém blob antigo;
- delete staged remove path;
- novo arquivo entra;
- parent é HEAD anterior;
- blobs inalterados são reutilizados;
- nenhuma cópia nova de blob idêntico é criada.

---

# 11. `init`

## Testes

- cria `.fit`;
- gera `repositoryId`;
- gera `peerId`;
- IDs são UUIDs válidos;
- IDs são diferentes;
- nome do repo é preservado;
- `formatVersion` existe;
- `HEAD` representa repo sem commits;
- index começa vazio;
- diretórios necessários existem.

## Edge cases

- repo já inicializado;
- `.fit` parcial;
- sem permissão de escrita;
- nome vazio;
- nome Unicode;
- restart mantém IDs;
- dois repos com mesmo nome têm IDs diferentes.

---

# 12. Primeira vertical slice de CLI

## Acceptance test

```bash
fit init projeto

echo hello > hello.txt

fit add hello.txt
fit commit -m "first"
```

Validar:

```text
CLI
↓
filesystem
↓
blob
↓
index
↓
commit
↓
HEAD
```

Ainda sem RabbitMQ.

---

# 13. `status`

## Estados básicos

```text
HEAD igual working tree + index vazio => CLEAN
working tree modificado => DIRTY
index modificado => DIRTY
arquivo tracked deletado => DIRTY
novo staged => DIRTY
```

## Edge cases

- dirty + transporte offline;
- merge em andamento;
- merge com conflito;
- behind;
- ahead;
- diverged;
- arquivo untracked;
- arquivo voltou exatamente ao conteúdo do HEAD;
- staged blob igual ao HEAD;
- staged delete de path já inexistente;
- mtime diferente com conteúdo igual;
- binários.

Timestamp não deve determinar dirty.

---

# 14. Checkout

## Happy path

```text
A -> B
checkout A
```

## Testes

- restaura conteúdo;
- remove arquivos ausentes no target;
- cria arquivos ausentes;
- move HEAD;
- index segue semântica definida.

## Edge cases

- working tree dirty;
- index dirty;
- target inexistente;
- commit corrompido;
- parent ausente;
- blob ausente;
- blob corrompido;
- todos os blobs são validados antes de alterar working tree;
- erro de materialização não move HEAD;
- path inseguro;
- checkout do próprio HEAD;
- arquivo vira diretório;
- diretório vira arquivo;
- filesystem read-only;
- conteúdo binário.

---

# 15. Histórico e DAG

## Cenário

```text
A <- B <- C
```

## Testes

- A ancestor de B;
- A ancestor de C;
- B ancestor de C;
- inversos false.

## Edge cases

- mesma origem/destino;
- commit inexistente;
- parent ausente;
- grafo profundo;
- merge commit;
- diamond;
- ciclo malicioso;
- parent duplicado.

## Invariante

```text
timestamp nunca participa da ancestralidade
```

Criar teste com timestamps invertidos.

---

# 16. Relação entre heads

Cobrir:

```text
igual                     → CLEAN
local ancestor remote     → BEHIND
remote ancestor local     → AHEAD
nenhum ancestor do outro  → DIVERGED
```

## Edge cases

- remote head desconhecido;
- parent remoto faltando;
- múltiplos remote heads;
- dois peers no mesmo head;
- remote ref muda de B para C;
- remote ref volta a ancestral;
- heads concorrentes.

---

# 17. Merge base

## Cenário

```text
A-B-C
   \
    D
```

```text
merge-base(C,D)=B
```

## Edge cases

- um head é ancestor do outro;
- raiz comum;
- histórias longas;
- merge commit;
- diamond;
- duas merge bases próximas;
- nenhum ancestral comum;
- commit ausente.

---

# 18. Fast-forward

## Testes

- local ancestor do target;
- HEAD muda;
- working tree vira target;
- nenhum commit novo.

## Edge cases

- target igual HEAD;
- target é ancestor do HEAD;
- target divergente;
- blob ausente;
- blob corrompido;
- working tree dirty;
- index dirty;
- merge em andamento;
- erro durante materialização.

---

# 19. 3-way merge de manifesto

Cobrir tabela de estados:

```text
base ours theirs

A    A    A      => A
A    B    A      => B
A    A    B      => B
A    B    B      => B
A    B    C      => merge/conflict

A    -    -      => delete
A    A    -      => delete theirs
A    -    A      => delete ours
A    B    -      => modify/delete
A    -    B      => modify/delete

-    B    -      => B
-    -    B      => B
-    B    B      => B
-    B    C      => add/add
```

## Edge cases

- muitos paths independentes;
- conflito em um path não bloqueia análise dos demais;
- conteúdo igual gera hash igual;
- rename é delete + add.

---

# 20. Merge textual

## Sem conflito

- mudanças em linhas diferentes;
- ours apenas;
- theirs apenas;
- inserções independentes;
- deleções independentes.

## Com conflito

- mesma linha alterada diferente;
- delete vs modify;
- inserção concorrente mesma posição;
- arquivo inteiro alterado por ambos.

## Edge cases

- arquivo vazio;
- uma linha sem newline final;
- CRLF vs LF;
- Unicode;
- UTF-8 multibyte;
- linha muito longa;
- muitas linhas;
- arquivo válido em UTF-8 mas “binário-like”.

---

# 21. Merge binário

## Testes

- apenas ours muda;
- apenas theirs muda;
- ambos geram mesmo blob;
- ambos alteram diferente;
- add/add igual;
- add/add diferente.

Nunca usar merge textual em binário.

---

# 22. Estado de conflito

## Testes

```text
merge produz conflito
MERGE_HEAD persistido
MERGE_BASE persistido
status = CONFLICTED
```

## Edge cases

- estado persistido antes de alterar working tree;
- restart preserva conflito;
- novo merge é rejeitado;
- checkout durante merge é rejeitado;
- resolução via `add`;
- resolução via `rm`;
- resolver parte mantém conflito;
- resolver tudo permite commit;
- merge commit tem `[ours, theirs]`;
- index preserva conflitos restantes.

---

# 23. `merge --abort`

## Cenário principal

```text
antes:
HEAD=A
index=X
working tree=W

merge
conflito

abort

HEAD=A
index=X
working tree=W
```

## Edge cases

- abort imediato;
- abort após editar arquivos;
- abort após resolver parcialmente;
- abort após restart;
- abort sem merge;
- arquivo originalmente inexistente volta a inexistir;
- arquivo originalmente existente volta aos bytes exatos;
- arquivo criado durante resolução;
- falha parcial no abort;
- blob necessário à restauração ausente/corrompido.

---

# 24. Atomicidade e crash safety

Fazer fault injection em:

```text
antes de temp write
durante temp write
depois de temp write
antes de rename
depois de rename
```

Aplicar a:

- blob;
- commit;
- index;
- HEAD;
- refs;
- MERGE_HEAD;
- MERGE_BASE;
- SYNC_HEADS.

## Invariantes

```text
nunca ler arquivo parcialmente escrito como válido
HEAD nunca aponta para commit inexistente
index sempre é JSON inteiro ou versão anterior
```

---

# 25. Lock local

## Testes concorrentes

- dois `commit`;
- `serve` + `commit`;
- `serve` + `sync`;
- `add` + `commit`;
- dois `serve`.

## Edge cases

- processo segurando lock morre;
- lock stale;
- liberação normal;
- segunda operação não corrompe estado;
- apenas um `serve` por peer.

---

# 26. Completude de objetos

## Regra

```text
commit presente
parent presente
todos blobs presentes
=> complete
```

## Edge cases

- commit presente, blob ausente;
- blob corrompido;
- parent ausente;
- grandparent ausente;
- blob de ancestor ausente;
- blob compartilhado;
- commit independente presente;
- manifest vazio.

---

# 27. Fetch recursivo sem rede

Usar fonte fake.

## Testes

- buscar head;
- buscar parent;
- buscar grandparent;
- buscar blobs;
- não buscar objetos válidos existentes;
- deduplicar blob;
- commit antes de parent;
- blob antes de commit;
- parent presente com blob faltando;
- commit inválido;
- blob inválido;
- not found;
- timeout;
- retry;
- segunda source válida após primeira inválida.

---

# 28. Mensagens e protocolo

## Envelope

Testar:

- `protocolVersion`;
- `messageId`;
- `type`;
- `repositoryId`;
- `senderPeerId`;
- `correlationId`;
- `sentAt`.

## Edge cases

- protocol version errada;
- repository ID errado;
- sender ID ausente;
- correlation ID incorreto;
- message type desconhecido;
- payload inválido;
- campo obrigatório ausente;
- JSON malformado;
- mensagem duplicada;
- `sentAt` não influencia causalidade.

---

# 29. Transporte fake hostil

O fake transport deve conseguir:

```text
drop
duplicate
delay
reorder
```

## Testes

- `head.announce` duplicado;
- `commit.response` duplicado;
- `blob.response` duplicado;
- resposta após timeout;
- resposta antes de dependência;
- duas fontes respondem;
- primeira inválida, segunda válida;
- primeira válida, segunda válida;
- correlation mismatch;
- request duplicado;
- reply-to inválido;
- peer desaparece no meio do request.

---

# 30. Developer peer ao receber anúncio

## Teste

```text
recebe head.announce
→ grava remote ref
```

## Negativos obrigatórios

```text
não baixa commit
não baixa blob
não move HEAD
não toca working tree
não toca index
```

## Edge cases

- mesmo head;
- novo head;
- ancestral antigo;
- head concorrente;
- repository ID errado;
- head desconhecido localmente.

---

# 31. `sync`

## Happy path

```text
A remote: A-B
B local:  A

sync B
```

Esperado:

```text
B baixa B
B possui closure completo
HEAD continua A
working tree continua A
index continua igual
```

## Edge cases

- tudo já presente;
- commit presente, blob faltando;
- child presente, parent faltando;
- blob compartilhado;
- remote head muda durante sync;
- novo head após snapshot fica para próximo sync;
- nenhum peer responde;
- uma source falha;
- várias sources;
- conteúdo inválido;
- timeout;
- retry limitado;
- sync interrompido;
- restart retoma;
- objeto válido não é baixado novamente;
- `SYNC_HEADS` pending;
- complete só com closure completo.

---

# 32. Interface de transporte

Só agora extrair uma interface.

Ela deve nascer do transporte fake já funcional e dos usos reais.

Possível forma:

```go
type Transport interface {
    Publish(...)
    Request(...)
    Subscribe(...)
}
```

Mas a interface final deve ser descoberta pelos testes.

Regra:

```text
FakeTransport continua passando sem mudança semântica
```

---

# 33. RabbitMQ adapter

## Contract tests

- cria e binda exchange;
- fila efêmera por processo;
- reply queue temporária;
- request chega a peers;
- response usa reply-to;
- correlation ID é preservado;
- blob usa `application/octet-stream`;
- blob não é base64;
- limite de mensagem gera erro explícito;
- disconnect remove fila efêmera;
- reconnect recria bindings;
- broker indisponível não afeta operações locais.

---

# 34. `serve`

## Testes

- responde `head.request`;
- responde `commit.request`;
- responde `blob.request`;
- anuncia presença/head ao iniciar;
- processa anúncios.

## Edge cases

- segundo `serve`;
- broker cai;
- broker volta;
- conexão cai;
- peer reinicia;
- refs persistem;
- HEAD persiste;
- `peerId` persiste;
- não depende de replay de mensagens.

---

# 35. Discovery

## Testes

- discover por ID;
- discover por nome;
- peer responde com offer;
- repo offline não responde.

## Edge cases

- mesmo nome, IDs diferentes;
- várias offers do mesmo repo;
- offer duplicada;
- repository ID errado;
- offer sem head;
- source desaparece depois da offer;
- nenhuma resposta;
- timeout;
- resposta atrasada.

---

# 36. Clone

## Primeiro cenário

```text
1 commit
1 blob
1 peer remoto
```

## Depois testar

- vários commits;
- merge commit;
- blobs compartilhados;
- blob vazio;
- binário;
- vários peers oferecendo;
- primeira source timeout;
- primeira source envia commit inválido;
- primeira source envia blob inválido;
- segunda funciona;
- parent chega depois;
- blob fora de ordem;
- clone interrompido;
- resume;
- nenhum peer;
- clone incompleto não parece válido;
- novo clone ganha `peerId`;
- mantém `repositoryId`;
- working tree só é materializado com closure válido.

---

# 37. Replica

## Happy path

```text
replica HEAD=A
remote head=B
A ancestor B
```

Esperado:

```text
download
validate
fast-forward automático
```

## Edge cases

- mesmo head;
- replica ahead;
- divergência;
- dois heads concorrentes;
- commit presente, blobs ausentes;
- conteúdo inválido;
- peer desaparece;
- broker restart;
- replica restart;
- merge commit remoto descendente permite FF;
- nunca cria merge commit;
- nunca resolve conflito;
- nunca escolhe silenciosamente entre heads divergentes.

---

# 38. Failure suite final

Executar E2E com processos reais.

## Cenários

```text
commit com RabbitMQ desligado
broker cai durante sync
broker cai durante clone
broker reinicia
peer fonte morre durante commit.request
peer fonte morre durante blob.request
peer destino morre durante sync
peer morre durante merge
peer reinicia durante conflito
mensagem perdida
mensagem duplicada
mensagens fora de ordem
resposta inválida antes da válida
download truncado
filesystem read-only
falha durante atomic rename
```

Para cada um validar:

```text
1. estado local continua válido?
2. nova execução consegue convergir?
```

---

# 39. Property tests e fuzzing

Fuzzar:

```go
func FuzzNormalizePath(f *testing.F)
func FuzzParseCommit(f *testing.F)
func FuzzParseIndex(f *testing.F)
func FuzzCanonicalCommit(f *testing.F)
func FuzzMergeText(f *testing.F)
func FuzzProtocolEnvelope(f *testing.F)
```

## Propriedades

```text
NormalizePath nunca escapa do root.

Canonicalize(x) é determinístico.

GetBlob(PutBlob(x)) == x.

CommitID(commit) == CommitID(decode(encode(commit))).

Merge(base, base, theirs) == theirs.

Merge(base, ours, base) == ours.

Parse de input arbitrário nunca causa panic.
```

---

# 40. Camadas de teste

## Unitários

Rodam constantemente:

```bash
go test ./...
```

Cobrem:

- hash;
- paths;
- index;
- commit canonicalization;
- DAG;
- merge;
- parser;
- protocolo.

## Integração

Usam:

- filesystem real;
- `t.TempDir()`;
- repos reais;
- fake transport;
- restart simulado.

## E2E

Usam:

- binário real;
- RabbitMQ real;
- múltiplos processos;
- kill/restart;
- falhas reais.

Não colocar RabbitMQ em teste que valida index, merge ou DAG.

---

# Ordem final de implementação

```text
BlobID
→ blob persistence/integrity
→ safe paths
→ add
→ index
→ rm
→ commit canonicalization
→ commit persistence
→ HEAD
→ snapshot completo
→ init
→ CLI vertical slice
→ status
→ checkout
→ log
→ ancestry
→ head relation
→ merge base
→ fast-forward
→ manifest merge
→ text merge
→ binary conflicts
→ merge state
→ abort
→ crash safety
→ lock
→ object completeness
→ recursive fetch
→ protocol messages
→ hostile fake transport
→ developer announcements
→ sync
→ sync resume
→ transport interface
→ RabbitMQ
→ serve
→ discovery
→ clone
→ replica
→ end-to-end failure suite
→ fuzzing pesado
```

---

# Critério para passar de uma etapa à próxima

Não avançar porque “a feature parece pronta”.

Avançar somente quando:

```text
happy path verde
+
edge cases conhecidos verdes
+
estado inválido rejeitado
+
falha não deixa corrupção
```

Quando um bug real aparecer:

```text
1. reproduzir com teste
2. ver RED
3. corrigir
4. ver GREEN
5. refatorar se necessário
```

Com o tempo, a suite de testes passa a ser a especificação executável do FIT.
