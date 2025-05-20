## Funis (Funnels)

### Listar todos os funis
- **Rota:** `GET /v1/funnels`
- **Requer:** Nada
- **Retorna:** Lista de funis
- **Exemplo de requisição:**
```bash
curl -X GET http://localhost:PORT/v1/funnels
```

### Obter um funil específico
- **Rota:** `GET /v1/funnels/{id}`
- **Requer:** Parâmetro de rota `id` (ObjectID do funil)
- **Retorna:** Funil detalhado
- **Exemplo de requisição:**
```bash
curl -X GET http://localhost:PORT/v1/funnels/ID_DO_FUNIL
```

### Criar um funil
- **Rota:** `POST /v1/funnels`
- **Requer:** JSON no corpo:
```json
{
  "name": "Nome do Funil",
  "type": "Tipo",
  "stages": [
    { "name": "Etapa 1", "related_leads": [], "related_budgets": [] }
  ]
}
```
- **Retorna:** 201 Created
- **Exemplo de requisição:**
```bash
curl -X POST http://localhost:PORT/v1/funnels \
  -H 'Content-Type: application/json' \
  -d '{"name":"Funil Teste","type":"vendas","stages":[{"name":"Prospecção"}]}'
```

### Atualizar um funil
- **Rota:** `PATCH /v1/funnels/{id}`
- **Requer:** Parâmetro de rota `id` e JSON no corpo (campos a atualizar)
- **Retorna:** 200 OK
- **Exemplo de requisição:**
```bash
curl -X PATCH http://localhost:PORT/v1/funnels/ID_DO_FUNIL \
  -H 'Content-Type: application/json' \
  -d '{"name":"Novo Nome"}'
```

### Deletar um funil
- **Rota:** `DELETE /v1/funnels/{id}`
- **Requer:** Parâmetro de rota `id`
- **Retorna:** 200 OK
- **Exemplo de requisição:**
```bash
curl -X DELETE http://localhost:PORT/v1/funnels/ID_DO_FUNIL
```

---

## Leads

### Listar todos os leads
- **Rota:** `GET /v1/leads`
- **Requer:** Parâmetros opcionais de paginação: `page`, `pageSize`
- **Retorna:** Lista paginada de leads
- **Exemplo de requisição:**
```bash
curl -X GET 'http://localhost:PORT/v1/leads?page=1&pageSize=10'
```

### Obter um lead específico
- **Rota:** `GET /v1/leads/{id}`
- **Requer:** Parâmetro de rota `id` (ObjectID do lead)
- **Retorna:** Lead detalhado
- **Exemplo de requisição:**
```bash
curl -X GET http://localhost:PORT/v1/leads/ID_DO_LEAD
```

### Criar um lead
- **Rota:** `POST /v1/leads`
- **Requer:** JSON no corpo:
```json
{
  "name": "Nome do Lead",
  "nickname": "Apelido",
  "phone": "Telefone",
  "type": "Tipo",
  "segment": "Segmento",
  "status": "Status",
  "source": "Origem",
  "unique_id": "ID Único",
  "related_budgets": [],
  "related_orders": [],
  "related_client": "ObjectID",
  "rating": "",
  "notes": "",
  "responsible": "ObjectID"
}
```
- **Retorna:** 201 Created
- **Exemplo de requisição:**
```bash
curl -X POST http://localhost:PORT/v1/leads \
  -H 'Content-Type: application/json' \
  -d '{"name":"Lead Teste","phone":"11999999999"}'
```

### Atualizar um lead
- **Rota:** `PATCH /v1/leads/{id}`
- **Requer:** Parâmetro de rota `id` e JSON no corpo (campos a atualizar)
- **Retorna:** 200 OK
- **Exemplo de requisição:**
```bash
curl -X PATCH http://localhost:PORT/v1/leads/ID_DO_LEAD \
  -H 'Content-Type: application/json' \
  -d '{"status":"Novo Status"}'
```

---

## Relatórios

### Listar relatórios (leads)
- **Rota:** `GET /v1/reports`
- **Requer:** Nada
- **Retorna:** Lista paginada de leads (igual ao endpoint `/v1/leads`)
- **Exemplo de requisição:**
```bash
curl -X GET http://localhost:PORT/v1/reports
```

---

> Substitua `PORT` pela porta configurada no seu ambiente.
> Substitua `ID_DO_FUNIL` e `ID_DO_LEAD` pelos respectivos ObjectIDs.
