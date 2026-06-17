# Go-Init — локальный Kubernetes (k3d + Ansible)

Развёртывание через **Ansible playbooks** в WSL / Linux.

## Требования

- WSL2 или Linux
- Docker (Docker Desktop + WSL integration)
- Ansible: `sudo apt install ansible` или `pip install ansible`

## Быстрый старт

```bash
cd ~/projects/Go-Init/deploy

ansible-playbook -i inventory/hosts.yml playbooks/setup.yml
```

> На `/mnt/c` или `/mnt/b` Ansible может игнорировать `ansible.cfg`.  
> Тогда явно укажи inventory: `-i inventory/hosts.yml` (роли лежат в `playbooks/roles/`).

В отдельном терминале — проброс портов:

```bash
ansible-playbook -i inventory/hosts.yml playbooks/port-forward.yml
```

## Плейбуки

| Плейбук | Назначение |
|---------|------------|
| `playbooks/setup.yml` | Полный подъём: tools → k3d → build → deploy |
| `playbooks/cluster.yml` | Только k3d-кластер |
| `playbooks/build.yml` | Сборка и импорт Docker-образов |
| `playbooks/deploy.yml` | `kubectl apply -k k8s/` |
| `playbooks/restart.yml` | Rollout restart приложений |
| `playbooks/port-forward.yml` | Проброс портов на localhost |
| `playbooks/teardown.yml` | Удалить ресурсы из k8s |
| `playbooks/teardown-all.yml` | Ресурсы + удалить кластер k3d |

## Endpoints

| Сервис | URL |
|--------|-----|
| Frontend | http://localhost:5173 |
| GraphQL | http://localhost:60013/graphql |
| Kafka UI | http://localhost:8082 |
| MinIO | http://localhost:9000 |

## Настройка

Переменные в `group_vars/all.yml`:

- `cluster_name`, `namespace`, `worker_nodes`
- `use_port_maps: true` — проброс портов через k3d (на Docker Desktop часто падает)
- `docker_images` — список образов для сборки

## Структура

```
deploy/
  ansible.cfg
  inventory/hosts.yml
  group_vars/all.yml
  playbooks/
    roles/         # ansible roles
  k8s/             # манифесты (kustomize)
```

## Обновление после изменений в коде

```bash
ansible-playbook -i inventory/hosts.yml playbooks/build.yml
ansible-playbook -i inventory/hosts.yml playbooks/restart.yml
```

## Диагностика

```bash
kubectl get pods -n go-init
```
