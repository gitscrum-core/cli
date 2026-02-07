# GitScrum CLI - Roadmap

> **Comando:** `gitscrum` (alias: `gs`)  
> **Linguagem:** Go  
> **Autenticação:** OAuth 2.0 Device Flow (como MCP)

> ⚠️ **BENCHMARK: SUPERAR LINEAR.APP**  
> Se não superar Linear CLI em UX, features e DX, **PRECISA SER REESCRITO**.

---

## Fase 0: Setup do Projeto ✅

- [x] Inicializar projeto Go (`go mod init github.com/gitscrum-team/cli`)
- [x] Configurar estrutura de diretórios
  - [x] `cmd/gitscrum/` - Entry point
  - [x] `pkg/cmd/` - Cobra commands (padrão GitHub CLI)
  - [x] `pkg/api/` - HTTP client
  - [x] `pkg/auth/` - OAuth Device Flow
  - [x] `pkg/config/` - Local config
  - [x] `pkg/output/` - Formatters (table, json)
  - [x] `pkg/git/` - Git repository detection
  - [x] `pkg/spinner/` - Loading indicators
- [x] Adicionar dependências
  - [x] `github.com/spf13/cobra` - CLI framework
  - [x] `github.com/spf13/viper` - Config management (complementa Cobra)
  - [x] `github.com/fatih/color` - Output colorido
  - [x] `github.com/pkg/browser` - Abrir browser
  - [x] `github.com/briandowns/spinner` - Spinner para operações >500ms
  - [x] `github.com/go-git/go-git/v5` - Git repository detection
- [x] Configurar GoReleaser (`.goreleaser.yaml`)
- [x] GitHub Actions workflow (`.github/workflows/release.yml`)
- [x] Shell completion files (bash, zsh, fish, powershell)
- [x] Criar README.md completo
- [x] Criar Makefile com targets de build/test/release
- [x] **BUILD TESTADO:** `bin/gitscrum.exe` funcionando ✓

---

## Fase 1: Core (v0.1) 🎯 - EM PROGRESSO

### Auth
- [x] `gitscrum login` - OAuth Device Flow (estrutura pronta, **PENDENTE: backend OAuth**)
- [x] `gitscrum logout` - Remove token local
- [x] `gitscrum whoami` - Mostra usuário autenticado
- [x] `gitscrum status` - Status da sessão

### UX Essenciais
- [x] `gitscrum init` - Setup inicial interativo (auth + config + detecta .git) ✓
- [x] Spinner para operações >500ms (`github.com/briandowns/spinner`)
- [ ] Cache local opcional para comandos list (JSON em `~/.gitscrum/cache/`)
- [x] `gitscrum config` - Ver config atual
- [x] `gitscrum config set workspace <slug>` - Define workspace padrão
- [x] `gitscrum config set project <slug>` - Define projeto padrão
- [ ] `gitscrum config alias` - Cria alias 'gs'
- [x] `gitscrum config reset` - Limpa config
- [x] Persistir config em `~/.gitscrum/config.yaml`

### Tasks
- [x] `gitscrum tasks` - Listar minhas tasks (**PENDENTE: API integração**)
- [x] `gitscrum tasks today` - Tasks para hoje ✓
- [x] `gitscrum tasks -p <project>` - Tasks de um projeto
- [x] `gitscrum tasks --filter blocker` - Filtrar blockers ✓
- [x] `gitscrum tasks --filter bug` - Filtrar bugs ✓
- [x] `gitscrum tasks --assignee @user` - Por assignee ✓
- [x] `gitscrum tasks view <code>` - Ver detalhes (GS-123)
- [x] `gitscrum tasks create "<title>" -p <project>` - Criar task
- [x] `gitscrum tasks update <code> --column="<status>"` - Atualizar ✓
- [x] `gitscrum tasks complete <code>` - Marcar como done ✓
- [x] `gitscrum tasks assign <code> @user` - Atribuir ✓

### Timer ⏱️
- [x] `gitscrum timer` - Ver timer ativo (estrutura pronta, **PENDENTE: API**)
- [x] `gitscrum timer start <code>` - Iniciar timer
- [x] `gitscrum timer start <code> -m "msg"` - Com descrição
- [x] `gitscrum timer stop` - Parar timer ativo
- [x] `gitscrum timer log <code> <duration>` - Log manual (2h30m)

### Output
- [x] Output padrão: tabelas coloridas
- [x] Flag `--json` para output JSON
- [x] Flag `-q` para quiet mode (só IDs)
- [x] Factory pattern para DI

---

## Fase 1.5: Git-Aware 🚀 (DIFERENCIAÇÃO CRÍTICA)

> **API já suporta:** GitHub, GitLab, Bitbucket  
> **Endpoints:** `/integrations/{provider}/branches`, `/integrations/{provider}/pull-requests`

### Detecção de Contexto
- [x] Detectar `.git` no diretório atual (`go-git/go-git`)
- [x] Extrair branch atual e remote origin
- [x] Parser de task code: `feature/GS-123-title` → `GS-123`
- [ ] Auto-resolver projeto GitScrum baseado em remote URL

### Comandos Git-Aware
- [x] `gitscrum tasks current` - Detecta branch, mostra task vinculada ✓
- [x] `gitscrum tasks branch <code>` - Cria branch no GitHub + checkout local ✓
  - API: `POST /integrations/github/branches/{issueUuid}`
  - Formato: `feature/{CODE}-{number}-{slug}` (ex: `feature/GS-123-fix-auth`)
- [x] `gitscrum tasks branches <code>` - Lista branches vinculadas à task ✓
  - API: `GET /integrations/github/branches/{issueUuid}`
- [x] `gitscrum tasks pr <code>` - Abre browser para criar PR ✓
  - URL: `https://github.com/{repo}/compare/{branch}?expand=1&title=...`
- [x] `gitscrum tasks prs <code>` - Lista PRs vinculados à task ✓
  - API: `GET /integrations/github/pull-requests/{issueUuid}`
- [x] `gitscrum tasks unlink-branch <uuid>` - Remove vínculo task ↔ branch ✓
  - API: `DELETE /integrations/github/branches/{branchUuid}`

### Git Hooks (Automação)
- [x] `gitscrum hooks install` - Instala hooks no repositório ✓
- [x] Hook `commit-msg`: Detecta task code, valida formato ✓
- [x] Hook `post-commit`: Atualiza status da task (opcional) ✓
- [x] Hook `pre-push`: Alerta sobre tasks não atribuídas ✓

### Inteligência de Commits
- [x] Detectar pattern `#TK-123` ou `GS-123` em commits ✓ (via hooks)
- [ ] Sugestão de mensagem: `git commit -m "feat(GS-123): ..."`

---

## Fase 2: Gestão Ágil (v0.2)

### Sprints
- [x] `gitscrum sprints` - Listar sprints ✓
- [x] `gitscrum sprints current` - Sprint ativo + KPIs ✓
- [x] `gitscrum sprints view <slug>` - Ver sprint específico ✓
- [x] `gitscrum sprints create "<title>" --start --end` - Criar ✓
- [x] `gitscrum sprints burndown` - Burndown chart (ASCII) ✓
- [x] `gitscrum sprints stats` - KPIs do sprint ✓

### Standup
- [x] `gitscrum standup` - Resumo: ontem, hoje, blockers ✓
- [x] `gitscrum standup completed` - O que foi feito ontem ✓
- [x] `gitscrum standup blockers` - Lista blockers ✓
- [x] `gitscrum standup team` - Status do time ✓
- [x] `gitscrum standup digest` - Digest semanal ✓
- [x] `gitscrum standup create` - Criar entrada de standup ✓

### Projects
- [x] `gitscrum projects` - Listar projetos ✓
- [x] `gitscrum projects view <slug>` - Ver detalhes ✓
- [x] `gitscrum projects stats` - Estatísticas ✓
- [x] `gitscrum projects create "<name>"` - Criar projeto ✓
- [x] `gitscrum projects members` - Ver membros ✓
- [x] `gitscrum projects switch <slug>` - Mudar projeto padrão ✓

### Workspaces
- [x] `gitscrum workspaces` - Listar workspaces ✓
- [x] `gitscrum workspaces view <slug>` - Ver detalhes ✓
- [x] `gitscrum workspaces stats` - Estatísticas ✓
- [x] `gitscrum workspaces switch <slug>` - Mudar workspace padrão ✓
- [x] `gitscrum workspaces members` - Ver membros ✓

### Timer Reports
- [x] `gitscrum timer report` - Relatório do dia (estrutura pronta)
- [x] `gitscrum timer report --week` - Relatório da semana ✓
- [x] `gitscrum timer report --team` - Relatório do time ✓
- [x] `gitscrum timer productivity` - Métricas de produtividade ✓

---

## Fase 3: Colaboração (v0.3)

### Tasks Avançado
- [x] `gitscrum tasks move <code> --to=<project>` - Mover task ✓
- [x] `gitscrum tasks duplicate <code>` - Duplicar ✓
- [x] `gitscrum tasks subtasks <code>` - Ver subtasks ✓

### Chat (Discussions)
- [x] `gitscrum chat` - Listar canais ✓
- [x] `gitscrum chat #<channel>` - Ver mensagens ✓
- [x] `gitscrum chat #<channel> "<msg>"` - Enviar mensagem ✓
- [x] `gitscrum chat unread` - Ver mensagens não lidas ✓
- [x] `gitscrum chat send` - Enviar mensagem (subcomando) ✓

### Wiki
- [x] `gitscrum wiki` - Listar páginas ✓
- [x] `gitscrum wiki view "<title>"` - Ver página ✓
- [x] `gitscrum wiki create "<title>" -f file.md` - Criar página ✓
- [x] `gitscrum wiki search "<query>"` - Buscar ✓
- [x] `gitscrum wiki edit "<slug>" -f file.md` - Editar página ✓

### Notifications
- [x] `gitscrum notifications` - Ver notificações ✓
- [x] `gitscrum notifications --unread` - Só não lidas ✓
- [x] `gitscrum notifications read <id>` - Marcar como lida ✓
- [x] `gitscrum notifications read-all` - Marcar todas como lidas ✓
- [x] `gitscrum notifications clear` - Limpar todas ✓

### Search
- [x] `gitscrum search "<query>"` - Busca global ✓

---

## Fase 4: ClientFlow CRM (v0.4)

### Clients
- [x] `gitscrum clients` - Listar clientes ✓
- [x] `gitscrum clients view "<name>"` - Ver cliente ✓
- [x] `gitscrum clients create "<name>"` - Criar cliente ✓
- [x] `gitscrum clients stats` - Estatísticas ✓
- [x] `gitscrum clients projects <slug>` - Ver projetos do cliente ✓

### Invoices
- [x] `gitscrum invoices` - Listar faturas ✓
- [x] `gitscrum invoices view <code>` - Ver fatura ✓
- [x] `gitscrum invoices create --client="<name>"` - Criar ✓
- [x] `gitscrum invoices send <code>` - Enviar ✓
- [x] `gitscrum invoices mark-paid <code>` - Marcar paga ✓

### Proposals
- [x] `gitscrum proposals` - Listar propostas ✓
- [x] `gitscrum proposals view <code>` - Ver proposta ✓
- [x] `gitscrum proposals create "<title>"` - Criar ✓
- [x] `gitscrum proposals send <code>` - Enviar ✓
- [x] `gitscrum proposals convert <code>` - Converter em projeto ✓

### Dashboard
- [x] `gitscrum crm` - Dashboard CRM ✓
- [x] `gitscrum crm revenue` - Pipeline de receita ✓
- [x] `gitscrum crm at-risk` - Clientes em risco ✓
- [x] `gitscrum crm pipeline` - Sales pipeline ✓

---

## Fase 5: Polish (v1.0)

### UX
- [ ] Completar `--help` para todos os comandos
- [ ] Adicionar exemplos em cada comando
- [ ] Mensagens de erro amigáveis
- [ ] Spinner/progress para operações longas
- [ ] Confirmação para ações destrutivas

### Integrações
- [ ] Git hooks templates
- [ ] GitHub Actions integration
- [ ] Shell completions (bash, zsh, fish, powershell)

### Distribuição
- [x] GoReleaser configurado (Homebrew, Scoop, deb/rpm)
- [x] GitHub Actions para releases automáticos
- [ ] Publicar Homebrew formula
- [ ] Publicar Scoop manifest (Windows)
- [ ] Publicar APT/YUM packages
- [ ] Binários pré-compilados (GitHub Releases)
- [ ] Documentação completa

---

## 🚨 GAPS CRÍTICOS (Developer Love) 🚨

### Webhooks (API já suporta - WebhookService)
> API: `GET/POST webhooks/resources` - 24 eventos disponíveis

- [ ] `gitscrum webhooks` - Listar webhooks configurados
- [ ] `gitscrum webhooks create <url> --events="issues.store,issues.update"` 
- [ ] `gitscrum webhooks test <id>` - Testar webhook (POST manual)
- [ ] `gitscrum webhooks delete <id>` - Remover webhook
- [ ] `gitscrum webhooks logs <id>` - Ver últimas entregas (WebhookLog)
- [ ] `gitscrum webhooks events` - Listar todos eventos disponíveis

**Eventos suportados pela API:**
```
issues.store, issues.update, issues.destroy, issues.move.board, issues.move.project
issues.assignees.store, issues.assignees.destroy
time-tracking.issues.start/stop/cancel/store/destroy
comments.issues.store/destroy, attachments.issues.store/destroy
user-stories.store/update/vote/destroy
sprints.store/update/destroy
```

### Aliases & Custom Commands
- [ ] `gitscrum alias` - Listar todos aliases
- [ ] `gitscrum alias set <name> <command>` - Criar alias
- [ ] `gitscrum alias remove <name>` - Remover alias
- [ ] Aliases armazenados em `~/.gitscrum/aliases.yaml`

**Exemplos pré-configurados:**
```
gs t       → gs tasks
gs today   → gs tasks --filter due:today --assignee @me
gs blocked → gs tasks --filter blocker
gs start   → gs timer start
gs stop    → gs timer stop
```

### Fast Feedback Loops (UX)
- [ ] `gitscrum doctor` - Valida config, connectivity, git setup
- [ ] Client-side validation antes de API call
- [ ] Sugestões quando comando falha: "Did you mean...?"
- [ ] `--dry-run` flag para comandos destrutivos
- [ ] Progress bar para uploads (wiki, attachments)

**Exemplo de output melhorado:**
```
$ gs tasks create "Fix bug"
❌ Error: Project required
💡 Tip: Set default project with: gs config set project <slug>
   Or use: gs tasks create "Fix bug" -p <project>

$ gs tasks create "Fix bug" -p api-backend
✓ Created GS-247 in 0.08s
  View: gs tasks view GS-247
  Branch: gs tasks branch GS-247
  Start: gs start GS-247
```

### Analytics & Metrics (API: WorkspaceAnalytics + ProjectAnalytics)
> API: `companies/reports/*`, `projects/{slug}/analytics/*`

- [ ] `gitscrum analytics cycle-time` - Média de tempo/task por tipo
- [ ] `gitscrum analytics throughput` - Tasks completadas/semana
- [ ] `gitscrum analytics blockers` - Tempo médio bloqueado
- [ ] `gitscrum analytics velocity` - Sprint velocity trend
- [ ] `gitscrum analytics workload` - Distribuição de tasks no time
- [ ] `gitscrum analytics cfd` - Cumulative Flow Diagram (ASCII)
- [ ] `gitscrum analytics pulse` - Manager pulse overview
- [ ] `gitscrum analytics risks` - Tasks/projects em risco
- [ ] Flag `--chart` para ASCII charts
- [ ] Flag `--me` para métricas pessoais (não surveillance)

**Exemplo:**
```
$ gs analytics cycle-time --me
📊 Your Average Cycle Time (Last 30 Days)
   Bug fixes: 2.3 hours
   Features: 8.7 hours
   Refactors: 4.1 hours
   
   💡 You're 15% faster than last month!
```

**Endpoints API disponíveis:**
```
GET companies/reports/cumulative-flow
GET companies/reports/project-age
GET companies/reports/weekly-activity
GET companies/manager-dashboard/pulse
GET companies/manager-dashboard/risks
GET projects/{slug}/analytics
GET projects/{slug}/analytics/cfd
GET projects/{slug}/analytics/timeline
GET time-trackings/analytics
```

---

## Estrutura Atual do Projeto ✅

```
cli/
├── cmd/gitscrum/
│   └── main.go               # Entry point
├── pkg/
│   ├── cmd/
│   │   ├── root/root.go      # Root command + version + completion
│   │   ├── auth/auth.go      # login, logout, whoami, status
│   │   ├── config/config.go  # set, get, reset
│   │   ├── tasks/tasks.go    # list, view, create, current, branch
│   │   ├── timer/timer.go    # start, stop, log, report
│   │   └── factory/factory.go # DI pattern
│   ├── api/client.go         # HTTP client + auth
│   ├── auth/
│   │   ├── token.go          # Token management
│   │   └── device.go         # OAuth Device Flow
│   ├── config/config.go      # Viper config management
│   ├── git/context.go        # Git detection + task code extraction
│   ├── output/formatter.go   # Table/JSON/Quiet formatters
│   └── spinner/spinner.go    # Loading indicators
├── .goreleaser.yaml          # Multi-platform releases
├── .github/workflows/
│   └── release.yml           # CI/CD
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── TODO.md
```

---

## Referências

- **MCP Server**: `c:\Users\renat\Projects\GitScrum\mcp` - Lógica de API
- **API Laravel**: `c:\Users\renat\Projects\GitScrum\api` - Backend
- **App Vue**: `c:\Users\renat\Projects\GitScrum\app` - Frontend
- **OAuth Device Flow**: RFC 8628
- **Cobra CLI**: https://github.com/spf13/cobra
- **GitHub CLI (gh)**: https://github.com/cli/cli - Arquitetura de referência

---

## Notas

1. **Reutilizar lógica do MCP**: Consultar `GitScrumClient.ts` para endpoints e estrutura de dados
2. **OAuth Device Flow**: Mesmo fluxo do MCP (`DeviceAuthenticator.ts`)
3. **Config local**: Armazenar em `~/.gitscrum/` (token, workspace, projeto)
4. **Alias 'gs'**: Criar symlink ou wrapper script durante instalação

---

## Próximos Passos

1. **Backend:** Implementar OAuth Device Flow no Laravel API
2. **CLI:** Integrar comandos tasks/timer com API real
3. **CLI:** Implementar auto-resolver projeto baseado em remote URL
4. **CLI:** Adicionar comandos git-aware restantes (branches, pr, prs, unlink-branch)
5. **CLI:** Adicionar testes unitários (`go test ./...`)
6. **Release:** Publicar v0.1-alpha no GitHub Releases
