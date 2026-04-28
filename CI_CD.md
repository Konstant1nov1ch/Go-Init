# CI/CD: пайплайны Go-Init


---

## Обзор архитектуры

```
Разработчик (ветка → MR/PR в main)
        │
        ▼
┌───────────────────┐       ┌────────────────────┐
│   GitLab CI       │   или │  GitHub Actions     │
│   (.gitlab-ci.yml)│       │  (.github/workflows)│
└─────────┬─────────┘       └──────────┬─────────┘
          │                            │
          ▼                            ▼
   stage: build                 job: build
   stage: lint                  jobs: lint (matrix ×4)
   stage: test                  jobs: test (matrix ×4)
          │                            │
          └────────────┬───────────────┘
                       ▼
            go-init-common, go-init-manager,
            go-init-generator, go-init-publisher
```

| Платформа    | Файл конфигурации              | Где смотреть результат        |
|--------------|--------------------------------|-------------------------------|
| GitLab       | `.gitlab-ci.yml`               | Merge Request → **Pipelines** |
| GitHub       | `.github/workflows/ci.yml`   | Репозиторий → **Actions**    |

| Job (GitLab) | Stage | Назначение |
|--------------|-------|------------|
| `build`      | build | `go build` для всех четырёх Go-модулей |
| `lint`       | lint  | `golangci-lint run ./...` по каждому модулю |
| `test`       | test  | `go test ./... -short` по каждому модулю |

Образ раннера: **`golang:1.23-bookworm`**. Линтер в CI: **golangci-lint v1.64.8**.

---

## Когда запускается пайплайн

| Событие | Условие |
|---------|---------|
| Merge Request | целевая ветка **`main`** |
| Push | ветка **`main`** |

Задаётся в `.gitlab-ci.yml` блоком `workflow.rules` и в GitHub workflow — `on: pull_request` / `push` для `main`.

---

## Запуск (локально, как в CI)

Из **корня репозитория** (пути как на раннере: `go-init-manager` использует `replace` на `../go-init-common`).

```bash
# Сборка
cd go-init-common && go build ./... && cd ..
cd go-init-manager && go build -o /dev/null ./cmd && cd ..
cd go-init-generator && go build ./... && cd ..
cd go-init-publisher && go build ./... && cd ..
```

```bash
# Тесты
for d in go-init-common go-init-manager go-init-generator go-init-publisher; do
  (cd "$d" && go test ./... -count=1 -short)
done
```

```bash
# Линт (нужен golangci-lint, см. https://golangci-lint.run/welcome/install/)
for d in go-init-common go-init-manager go-init-generator go-init-publisher; do
  (cd "$d" && golangci-lint run ./...)
done
```

Сборка Docker-образа менеджера (как в проде с монорепо):

```bash
cd virtualization
docker compose -f docker-compose-all.yml build go_init_manager
```

---

## GitLab: настройка проекта

1. Файл **`.gitlab-ci.yml`** лежит в корне — закоммитить и запушить.
2. **Settings → CI/CD**: при необходимости включить **Merge request pipelines**.
3. **Runners**: для GitLab.com обычно достаточно shared runners; для self-hosted — зарегистрировать runner (Docker executor).
4. Открыть **Merge Request в `main`** → вкладка **Pipelines** → джобы `build`, `lint`, `test`.

Подробные шаги (защита веток, merge только после green pipeline): см. раздел ниже и [документацию GitLab](https://docs.gitlab.com/ee/ci/).

---

## GitHub Actions

Файл `.github/workflows/ci.yml`: при push/PR в **`main`** запускаются jobs **build**, **lint** (matrix по модулям), **test** (matrix). Просмотр: вкладка **Actions** в репозитории.

Если репозиторий только на GitLab, каталог `.github` можно не использовать.

---

## Структура файлов

```
Go-Init/
├── .gitlab-ci.yml                 ← GitLab: stages build → lint → test
├── CI_CD.md                       ← этот документ
├── docs/
│   └── CI_CD.md                   ← краткая ссылка на корневой отчёт
└── .github/
    └── workflows/
        └── ci.yml                 ← GitHub Actions
```

---

## Остановка / отключение

CI не создаёт долгоживущих сервисов. Чтобы **не запускать** пайплайны на MR в main — измените `workflow.rules` в `.gitlab-ci.yml` или отключите CI в **Settings → General → Visibility** (крайний случай).

---

## Полезные ссылки

| Тема | Ссылка |
|------|--------|
| GitLab CI YAML | https://docs.gitlab.com/ee/ci/yaml/ |
| Workflow rules | https://docs.gitlab.com/ee/ci/yaml/workflow.html |
| GitHub Actions | https://docs.github.com/en/actions |
| golangci-lint | https://golangci-lint.run/ |
