# Деплой app4every на Yandex Cloud VM

## IP сервера: `217.198.171.179`
## Домен: `jaronimo.work.gd`

---

## Шаг 1 — Убедись что DNS настроен

В панели управления доменом `jaronimo.work.gd` должна быть A-запись:

| Тип | Имя | Значение |
|-----|-----|----------|
| A | @ | 217.198.171.179 |

Проверь с локального ПК:
```bash
nslookup jaronimo.work.gd
# Должен вернуть: 217.198.171.179
```

---

## Шаг 2 — Yandex Cloud: открыть порты в Security Group

Перейди в **Yandex Cloud Console → VPC → Security Groups**.

Добавь **входящие правила**:

| Протокол | Порт | Источник | Зачем |
|----------|------|----------|-------|
| TCP | 80 | 0.0.0.0/0 | Let's Encrypt ACME + HTTP→HTTPS редирект |
| TCP | 443 | 0.0.0.0/0 | HTTPS трафик |
| UDP | 443 | 0.0.0.0/0 | HTTP/3 QUIC — опционально |

Проверь firewall на ВМ (если Ubuntu/Debian):
```bash
sudo ufw status
# Если активен:
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
```

---

## Шаг 3 — Pull и запуск на ВМ

```bash
ssh ubuntu@217.198.171.179

cd ~/app4every   # путь к проекту

git pull

# Пересобери Caddy с новым конфигом
docker compose up -d --force-recreate caddy

# Смотри логи — Caddy получит сертификат от Let's Encrypt
docker compose logs -f caddy
```

**Успешный результат в логах:**
```
{"level":"info","msg":"certificate obtained successfully","identifier":"jaronimo.work.gd"}
{"level":"info","msg":"Serving HTTPS"}
```

---

## Шаг 4 — Проверка

```bash
# Открой в браузере:
https://jaronimo.work.gd

# Или через curl:
curl -I https://jaronimo.work.gd
# Ожидаемо: HTTP/2 200
```

---

## Возможные проблемы

### ❌ Caddy не получает сертификат (timeout в логах)
- Проверь что порт 80 открыт снаружи: `curl http://217.198.171.179` с другого ПК
- DNS уже указывает на IP? `nslookup jaronimo.work.gd`
- Подожди — бесплатные DNS могут обновляться до часа

### ❌ Старый сертификат / ошибка TLS
```bash
docker compose down caddy
docker volume rm app4every_caddy_data
docker compose up -d caddy
```

### ❌ Vue Router — 404 при обновлении страницы
Добавь в Caddyfile перед последним `handle {}`:
```
@spa {
    not path /api/*
    file {
        try_files {path} /index.html
    }
}
handle @spa {
    reverse_proxy frontend:5173
}
```

---

## Архитектура

```
jaronimo.work.gd (217.198.171.179)
    │
    ├── :80  → Caddy → HTTP→HTTPS редирект
    └── :443 → Caddy (TLS: Let's Encrypt, автообновление)
                  │
                  ├── /api/v1/reviews*    → reviews-service:8082
                  ├── /api/v1/groups*     → reviews-service:8082
                  ├── /api/v1/screenshare* → screenshare-service:8083
                  ├── /api/v1/notes*      → notebook-service:8081
                  ├── /api/*              → auth-service:8080
                  └── /*                 → frontend:5173
```

---

## После успешного деплоя — смени секреты в docker-compose.yml!

```bash
# Сгенерировать JWT_SECRET:
openssl rand -hex 32

# Сгенерировать пароль БД:
openssl rand -base64 24
```
