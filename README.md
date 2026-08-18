# Core Service — Event Sourcing & CQRS Order Microservice

Go tabanlı, **Event Sourcing + CQRS + DDD** mimarisiyle tasarlanmış sipariş (order) mikroservisi. Yazma tarafı komutlar ile **KurrentDB** (Event StoreDB) üzerine domain event'leri yazar; okuma tarafında **PostgreSQL** ve **Elasticsearch** olmak üzere iki ayrı read model, persistent subscription'lar üzerinden asenkron olarak güncellenir.

## Özellikler

- **Event Sourcing**: `OrderAggregate` state machine kurallarını uygular, her geçerli komut bir domain event üretir (`order-{uuid}` stream'i)
- **CQRS**: Komut (yazma) ve sorgu (okuma) yolları tamamen ayrı has `OrderCommands` / `OrderQueries` facade'leri üzerinden yürütülür
- **Çift projeksiyon**: PostgreSQL ve Elasticsearch read modelleri, KurrentDB `$all` persistent subscription'ları ile beslenir
- **Optimistik eşzamanlılık**: Aggregate kaydedilirken `expectedRevision` kontrolü (yeni stream → `NoStream`, mevcut → son revision)
- **Uçtan uca OpenTelemetry tracing**: HTTP → komut handler → aggregate → event store → projeksiyon zincirinde trace context event metadata üzerinden taşınır
- **Çift transport**: Aynı handler'lar hem REST (Fiber) hem gRPC üzerinden sunulur (Swagger dahil)
- **Kapsamlı observability**: Tempo (traces), Prometheus (metrics), Loki + Promtail (logs), Grafana (dashboard)

## Mimari

```
                    ┌─────────────────────────────────────────────────┐
                    │                  Clients                        │
                    └──────────────┬──────────────────┬───────────────┘
                                   │                  │
                         HTTP :5010 │                  │ gRPC :5001
                                   ▼                  ▼
                    ┌─────────────────────────────────────────────┐
                    │        OrderHandlers (Fiber)   OrderGrpc    │
                    └─────────────────────┬───────────────────────┘
                                          │
                    ┌─────────────────────▼───────────────────────┐
                    │  OrderCommands (write)   OrderQueries (read)│
                    └──────┬───────────────────────┬──────────────┘
                           │                       │
              Handle + Apply                        │
                           ▼                       │
              ┌───────────────────────┐            │
              │  OrderAggregate       │            │
              │  (state machine)      │            │
              └──────────┬────────────┘            │
                         │ append                   │
                         ▼                          │
              ┌───────────────────────┐             │
              │  KurrentDB            │             │
              │  streams: order-{id}  │             │
              └──────────┬────────────┘             │
                         │ persistent subscriptions │
                         ▼ (fan-out)                │
   ┌────────────────────────┴──────────────────┐    │
   ▼                                           ▼    │
┌──────────────┐                      ┌───────────────────┐
│  Postgres    │                      │  Elasticsearch    │
│  projection  │                      │  projection       │
└──────┬───────┘                      └─────────┬─────────┘
       │ GetById / GetAll                        │ Search / GetByID
       ▼                                         ▼
  ┌───────────────────────────────────────────────────────┐
  │         OrderQueries (read models)                    │
  └───────────────────────────────────────────────────────┘
```

## Teknoloji Stack

| Katman | Teknoloji |
|---|---|
| Dil | Go 1.25 |
| HTTP | gofiber/fiber v2 |
| gRPC | google.golang.org/grpc + protobuf |
| Event Store | KurrentDB (KurrentDB-Client-Go v1.4.1) |
| Veritabanı (read model) | PostgreSQL (jackc/pgx v4) |
| Arama | Elasticsearch 7.17.29 (olivere/elastic v7) |
| Konfigürasyon | spf13/viper (YAML + env override) |
| Logging | uber-go/zap |
| Tracing | OpenTelemetry OTLP gRPC exporter |
| Metrics | prometheus/client_golang + go-grpc-prometheus |
| Health Check | heptiolabs/healthcheck |
| Swagger | swaggo/swag + fiber-swagger |

## Dizin Yapısı

```
core-service/
├── cmd/main.go                  # Giriş noktası: bağımlılık enjeksiyonu + sunucu başlatma
├── config/                      # Viper config yükleme + config.yaml
├── docs/                        # Swagger (docs.go, swagger.json/yaml)
├── proto/                       # order.proto + üretilmiş gRPC kodları
├── pkg/                         # Tekrar kullanılabilir paketler
│   ├── es/                      # Event sourcing çekirdeği (AggregateBase, AggregateStore)
│   │   └── store/               # KurrentDB tabanlı AggregateStore/EventStore
│   ├── eventstroredb/           # KurrentDB client fabrikası
│   ├── databases/               # PostgreSQL + Elasticsearch bağlantıları
│   ├── server/                  # HTTP (Fiber), gRPC, health check sunucuları
│   ├── middlewares/trace/       # OTel tracer provider + Fiber middleware
│   ├── trace/                   # OTLP exporter inicializasyonu
│   ├── logger/                  # Zap tabanlı logger arayüzü
│   ├── http_errors/ / grpc_errors/   # Hata eşleme yardımcıları
│   └── utils/                   # pagination, common yardımcılar
└── internal/
    ├── order/
    │   ├── aggregate/           # OrderAggregate + komut metotları + domain hataları
    │   ├── commands/v1/         # 7 komut + handler'ları (yazma yolu)
    │   ├── queries/             # GetOrderByID, Search (okuma yolu)
    │   ├── events/v1/           # V1_* event tipleri + constructor'lar
    │   ├── domain/              # Order, Payment, ShopItem modelleri + repo arayüzleri
    │   ├── projection/
    │   │   ├── postgres_projection/    # PG read model projeksiyonu
    │   │   └── elastic_projection/     # ES read model projeksiyonu
    │   ├── repository/          # PostgreSQL + Elasticsearch repo implementasyonları
    │   ├── delivery/            # HTTP (Fiber) + gRPC transport katmanları
    │   └── service/             # Commands + Queries facade'leri
    ├── dto/                     # HTTP request/response DTO'ları
    ├── mappers/                 # Domain ↔ DTO ↔ Proto dönüşümleri
    └── metrics/                 # Prometheus metrik sayaçları
```

## Kurulum & Çalıştırma

### Prerequisites

- Docker + Docker Compose
- Go 1.25+ (lokal geliştirme için)

### Tüm stack'i ayağa kaldırma

```bash
docker compose -f docker-compose.local.yaml up -d --build
```

Bu komut şu servisleri başlatır:

| Servis | Port | Açıklama |
|---|---|---|
| `server1` | 5010 (HTTP), 5001 (gRPC) | Mikroservis |
| `kurrentdb.db` | 2113 | Event store (eventsourcingdb) |
| `postgresql` | 5432 | Read model (PostgreSQL 16) |
| `node01` | 9200 | Elasticsearch 7.17.29 |
| `kibana` | 5601 | Elasticsearch UI |
| `otel-collector` | 4317 | OTel collector (→ Tempo) |
| `tempo` | 3200 | Trace backend |
| `prometheus` | 9090 | Metrics |
| `loki` | 3100 | Log storage |
| `promtail` | — | Docker loglarını Loki'ye gönderir |
| `grafana` | 3000 | Dashboard (anonim erişim) |
| `node_exporter` | 9101 | Host metrikleri |

## Konfigürasyon

Varsayılan ayarlar `config/config.yaml` içindedir. Öncelik sırası:

1. `-config` flag
2. `CONFIG_PATH` env değişkeni
3. `./config/config.yaml` (varsayılan)

Env override'ları:

```bash
GRPC_PORT=:5001
EVENT_STORE_CONNECTION_STRING=esdb://kurrentdb.db:2113?tls=false
ELASTIC_URL=http://node01:9200
```

## API

### REST (Fiber)

Base path: `/api/v1` — Swagger: `http://localhost:5010/doc/`

| Metot | Path | Açıklama |
|---|---|---|
| POST | `/api/v1/orders` | Yeni sipariş oluştur |
| PUT | `/api/v1/orders/pay/:id` | Siparişi öde |
| PUT | `/api/v1/orders/submit/:id` | Siparişi gönder (yalnızca ödenmiş) |
| POST | `/api/v1/orders/cancel/:id` | Siparişi iptal et (neden zorunlu) |
| POST | `/api/v1/orders/complete/:id` | Siparişi tamamla |
| PUT | `/api/v1/orders/address/:id` | Teslimat adresini değiştir |
| PUT | `/api/v1/orders/cart/:id` | Sepeti güncelle |
| GET | `/api/v1/orders/:id` | Sipariş detayı |
| GET | `/api/v1/orders/search` | Full-text arama (Elasticsearch) |

### gRPC

- Port: `:5001`, reflection geliştirme modunda açık
- Servis: `orderService` — 9 unary RPC: `CreateOrder`, `PayOrder`, `SubmitOrder`, `UpdateShoppingCart`, `CancelOrder`, `CompleteOrder`, `ChangeDeliveryAddress`, `GetOrderByID`, `Search`
- Proto: `proto/order.proto`

## Domain Model

Sipariş yaşam döngüsü (`OrderAggregate` state machine):

```
Created → Paid → Submitted → Completed
   │        │                   │
   │        └── Canceled        │
   └────────────── Canceled ←───┘
```

Kurallar:
- Ödeme yalnızca `Created` durumunda yapılabilir
- Gönderim yalnızca ödenmiş siparişte yapılabilir
- İptal için neden (reason) zorunludur; tamamlanmış sipariş iptal edilemez
- Adres değişikliği tamamlanmış siparişte yapılamaz

Event tipleri: `V1_ORDER_CREATED`, `V1_ORDER_PAID`, `V1_ORDER_SUBMITTED`, `V1_ORDER_COMPLETED`, `V1_ORDER_CANCELED`, `V1_SHOPPING_CART_UPDATED`, `V1_DELIVERY_ADDRESS_CHANGED`

## Observability

- **Traces**: Uygulama OTLP gRPC ile `otel-collector:4317`'e span gönderir → Tempo'ya iletir. Tempo metrics_generator servis-graflarını Prometheus'a remote write eder.
- **Logs**: Promtail, Docker container loglarını Loki'ye taşır.
- **Metrics**: Uygulama `core_service_*` counters + gRPC metrikleri (port `:8001`), Prometheus'a bağlı.
- **Grafana**: `http://localhost:3000` (anonim), Tempo TraceQL + Prometheus + Loki datasource'ları hazır tanımlı.

## Sağlık Kontrolü

- `/ready` (readiness) — PostgreSQL + Elasticsearch ping | port `:3001`
- `/live` (liveness) — port `:3001`

## Bilinen Eksikler / TODO'lar

- `PostgresRepository`'de `UpdateOrder`, `UpdateCancel`, `UpdatePayment`, `Complete`, `UpdateDeliveryAddress`, `UpdateSubmit`, `Delete` yöntemleri `panic("implement me")` — yalnızca `Create` ve `GetById` çalışır
- `public.orders` tablosu için migration dosyaları yok
- Swagger: `GET /orders/search` rotası `/:id` rotasından sonra kayıtlı olduğundan Fiber'de çakışma riski var
- Bazı handler'lar query parametrelerini `ctx.Query` yerine `ctx.Params` ile okuyor
- gRPC `Search` RPC'i yanlış metrik counter'ı artırıyor
- Doğrulama (`validator`) handler'larda yorum satırına alınmış
- `mongo` config anahtarları ve bağımlılıkları (template kalıntısı) kullanılmıyor
- `tempo-data/` runtime dosyaları repoya commit edilmiş durumda — `.gitignore`'a eklenmeli
- Test dosyaları yok (yalnızca `go-sqlmock` bağımlılığı mevcut)