# 02 — AWS-first integration, provider-agnostic core

Trạng thái: **chốt hướng thiết kế**. Phase 1 chỉ implement adapter đang cần.

ADR: [decisions/0001-aws-first-provider-agnostic.md](decisions/0001-aws-first-provider-agnostic.md)

## 1. Nguyên tắc

```text
Core (domain + use case)     không biết AWS / MinIO / DynamoDB
Port (interface)             hình dạng ưu tiên khả năng AWS
Adapter                      gắn core vào một nhà cung cấp cụ thể
Wiring                       chọn adapter lúc start, qua config
```

**AWS-first** nghĩa là: cổng được thiết kế để AWS dùng **tự nhiên** (presigned URL, object key mờ, IAM ở tầng infra). Không cắt cụt về “mẫu số chung nhỏ nhất” của mọi cloud rồi làm AWS trở nên vụng.

**Provider-agnostic core** nghĩa là: service/repository **không** import AWS SDK, không nhận `s3.Client`, không chứa SQL hay DynamoDB attribute map. Core nói ngôn ngữ domain: `CreateUploadURL`, `GetFile`, `ListChildren`.

**Thiết kế sẵn ≠ implement sẵn.** Spec ghi cổng + mapping. Code adapter chỉ xuất hiện khi phase cần.

## 2. RDS không phải adapter khác Postgres

Đây là chỗ dễ nhầm.

| Thứ | Là gì | Ảnh hưởng code domain |
|---|---|---|
| PostgreSQL local (Docker) | Cùng engine, cùng adapter | Không |
| Amazon RDS PostgreSQL | Postgres do AWS quản lý (host, backup, IAM) | Không |
| Amazon DynamoDB | Database NoSQL khác model | Có — adapter mới |

```text
MetadataStore port
        │
        ├── postgres adapter  ← local Docker VÀ RDS dùng chung
        │         local:  host=postgres:5432
        │         RDS:    host=xxx.rds.amazonaws.com
        │
        └── dynamodb adapter  ← thiết kế mapping, CHƯA implement
```

RDS là quyết định **deploy / Terraform**. DynamoDB là quyết định **data model**.

## 3. Catalog cổng

Mỗi cổng: một interface, một adapter production AWS, một adapter local nếu cần. Adapter thứ ba chỉ ghi spec.

### 3.1 ObjectStore — object storage (kho object)

AWS-first: **presigned URL là thao tác hạng nhất**, không phải afterthought.

```text
CreateUploadURL
CreateDownloadURL
Head / Stat
Delete
```

| | Adapter | Phase |
|---|---|---|
| Production | AWS S3 | Phase 1 |
| Local | MinIO (tương thích S3 API) | Phase 1 |
| Khác | GCS / Azure Blob | Chỉ thiết kế; không code |

Vì MinIO nói S3 API, Phase 1 **một implementation** `S3ObjectStore`, khác nhau ở endpoint/credential/bucket. Không viết hai codebase `S3` và `MinIO` trừ khi MinIO lệch hành vi thật sự.

Domain chỉ giữ `object_key` dạng chuỗi mờ (`users/{userId}/files/{fileId}`). Không nhét `s3://` hay region vào model nếu không cần.

Kiểu trả về là DTO của Filnest (`URL`, `ExpiresAt`), không phải type AWS SDK.

### 3.2 MetadataStore — metadata

Không một “god interface” khổng lồ. Cổng nằm ở repository theo domain, đúng modular monolith:

```text
UserRepository
FolderRepository
FileRepository
UploadSessionRepository
…
```

Service gọi repository. Không gọi `database/sql` hay DynamoDB từ handler/service.

| | Adapter | Phase |
|---|---|---|
| Mặc định | `postgres` (local hoặc RDS) | Phase 1 local; Phase 2 RDS |
| Học / Always Free | `dynamodb` | Thiết kế mapping; **không code Phase 1** |
| Khác | — | Không cam kết |

PostgreSQL vẫn là **source of truth mặc định** của sản phẩm. DynamoDB không thay Postgres trong Phase 1–2.

### 3.3 JobQueue — hàng đợi việc nền

```text
Enqueue
(consume nằm ở worker, không nhất thiết cùng process)
```

| | Adapter | Phase |
|---|---|---|
| Production | SQS (+ DLQ — Dead Letter Queue, hàng chữ lỗi) | Phase 3 |
| Local | In-memory hoặc queue trong process | Khi làm async |
| Khác | — | Không cam kết |

Phase 1: **không** có queue. Vẫn không được để service gọi SQS SDK trực tiếp khi tới Phase 3 — đi qua cổng.

### 3.4 Worker runtime — chỗ chạy việc nền

Concept chưa khóa Lambda vs ECS. Spec cũng không khóa.

| | Ý nghĩa | Phase |
|---|---|---|
| Local | Process Go `workers/file-processor` | Khi làm async |
| AWS | Lambda **hoặc** ECS | Phase 3+; chọn lúc deploy |
| Core | Job handler thuần Go, không biết Lambda event | Thiết kế sẵn |

Cổng ở đây là: **cùng một hàm xử lý job**. Adapter chỉ là cách đưa message vào hàm đó (SQS trigger Lambda, hay worker long-poll).

### 3.5 Auth

Phase 1: tự xây email/password + JWT. Đó là adapter `localjwt`.

Cognito (nếu sau này học): adapter `cognito`, **không** nhét vào Phase 1. Core chỉ thấy `UserID`, session, token claims đã được dịch sang kiểu domain.

### 3.6 Mailer — gửi mã verify

```text
Send(to, subject, body)  // body chứa mã; console adapter được phép log mã
```

| | Adapter | Phase |
|---|---|---|
| Không SMTP | `console` — in mã ra log/stdout | Phase 1 |
| Có SMTP env | `smtp` | Phase 1 |
| AWS | `ses` | Thiết kế; chưa code |

Wiring: SMTP đủ config → `smtp`, không thì `console`. Không gửi hai kênh cùng lúc. Core auth không biết SMTP hay console.

### 3.7 Observability và config

| Cổng | AWS-first | Local | Phase |
|---|---|---|---|
| Logs / metrics | CloudWatch | stdout / slog | Log structured từ đầu; CloudWatch khi deploy |
| Config / secrets | SSM / Secrets Manager | env / `.env` | Env Phase 1; AWS secrets khi deploy |
| IAM | Policy tối thiểu trên role | MinIO policy / dummy | Infra, không phải code domain |

IAM không phải interface Go. Nó là ranh giới quyền của adapter AWS.

## 4. DynamoDB — thiết kế sẵn, chưa implement

Mục đích ghi mapping ngay: khi học DynamoDB, biết chỗ cắm; không biến domain thành bảng quan hệ rồi “dịch sau”.

### 4.1 Việc DynamoDB làm tốt

Các bounded context (vùng nghiệp vụ khép) ít join, key rõ:

```text
upload_sessions
device presence / last seen
rate limit
idempotency key
cache trạng thái ngắn
```

### 4.2 Việc Postgres giữ

```text
folder tree lồng nhau
ownership + (sau này) ACL sharing
quota update kèm tạo file (transaction)
search tên file / FTS
reporting, dashboard, audit truy vấn linh hoạt
```

### 4.3 Sketch mapping (không phải schema chốt)

Chỉ để hình dung adapter tương lai. **Không** tạo bảng DynamoDB trong Phase 1.

```text
PK                          SK
USER#{userId}               PROFILE
USER#{userId}               FOLDER#{folderId}
USER#{userId}               FILE#{fileId}

GSI-1  (list theo parent):
  parentKey = USER#{userId}#PARENT#{folderId}
  SK        = NAME#{name}#FILE#{fileId}
```

Hệ quả đã biết trước:

- List con theo folder: được, nhờ GSI.
- Rename: update item; object storage không đụng — giống Postgres.
- Tree sâu / move subtree: đắt hơn Postgres.
- Full-text search: **không** có sẵn như Postgres FTS; sau này mới nghĩ OpenSearch nếu thật sự cần.
- Quota: cần TransactWrite; vẫn làm được nhưng khác tay Postgres `UPDATE users SET storage_used`.

Vì những hệ quả đó: DynamoDB là **adapter tùy chọn / học tập**, không phải đường mặc định của sản phẩm.

### 4.4 Config dành chỗ

Schema config **có khóa** cho metadata backend. Phase 1 chỉ nhận `postgres`. Giá trị `dynamodb` fail fast với lỗi rõ, chưa có implementation.

```text
FILNEST_OBJECT_STORE=s3          # s3 (MinIO hoặc AWS, khác endpoint)
FILNEST_METADATA_STORE=postgres  # postgres | dynamodb (dynamodb: chưa có)
FILNEST_QUEUE=none               # none | memory | sqs
FILNEST_INVITE_CODE=             # bắt buộc để register
FILNEST_MAILER=auto              # auto: smtp nếu có env, không thì console
```

Không tạo factory khổng lồ. Một hàm wiring trong `cmd/api`.

## 5. Điều cấm trong core

Không xuất hiện trong `internal/{file,folder,user,...}` service:

```text
github.com/aws/aws-sdk-go-v2/...
s3.Client, dynamodb.Client, sqs.Client
database/sql trong handler
SELECT ... viết thẳng trong use case
types.ObjectCannedACL, attributevalue.MarshalMap
```

Được:

```text
object_key string
FileID, UserID kiểu domain
PresignedURL{ URL, ExpiresAt }
error domain (NotFound, QuotaExceeded, ...)
```

SQL sống trong `postgres` adapter. DynamoDB marshal sống trong `dynamodb` adapter (khi có).

## 6. Chỗ đặt code (khi implement)

Bám monorepo concept §59, thêm chỗ adapter rõ ràng:

```text
apps/api/
  cmd/api/                 ← wiring / composition root
  internal/
    file/                  ← handler, service, port (interface)
    folder/
    user/
    platform/
      config/
      postgres/            ← Phase 1
      s3/                  ← Phase 1 (AWS + MinIO)
      mailer/              ← Phase 1 console + smtp
      dynamodb/            ← thư mục trống hoặc chưa tạo, đến khi làm
      sqs/                 ← Phase 3
```

Không tạo sẵn package rỗng cho mọi AWS service. Spec này **là** chỗ dành sẵn.

`ObjectStorage` trong concept §60 giữ nguyên ý; đổi tên thống nhất thành `ObjectStore` khi code để khớp catalog cổng.

## 7. Ma trận “làm gì / khi nào”

| Năng lực | Cổng trong spec | Code Phase 1 | Code sau |
|---|---|---|---|
| Object storage | Có | S3 + MinIO (một adapter) | — |
| Metadata Postgres | Có | Docker Postgres | RDS: đổi host/Terraform |
| Metadata DynamoDB | Mapping §4 | Không | Khi chủ động học NoSQL |
| Mailer | Có | console và/hoặc smtp | SES sau |
| Queue | Có | `none` | SQS Phase 3 |
| Worker | Job handler thuần | Không | Lambda hoặc process Go |
| Cognito | Ghi nhận | Không | Tùy chọn |
| CloudFront | Infra, không cổng app | Không | Phase CDN |
| Multi-cloud GCS/Azure | Ghi nhận | Không | Không cam kết |

Concept §23 “❌ Multi-storage providers” vẫn đúng: Phase 1 **không** ship nhiều cloud. Spec này chỉ siết: **cổng có từ đầu**, để sau không phải xé service.

## 8. AWS-first áp vào upload (ví dụ)

Flow concept §7 giữ nguyên. Phân lớp:

```text
Vue
  → Go handler
      → file service          (core: tạo session, kiểm tra quota)
          → FileRepository    (port → postgres)
          → ObjectStore       (port → s3 presign)
  → Vue PUT presigned URL     (thẳng S3 hoặc MinIO)
  → Go confirm
      → ObjectStore.Head
      → FileRepository.Update status
```

`file service` không biết URL là S3 hay MinIO. Adapter S3 biết. Đó là mức trừu tượng đủ dùng — không thêm layer “storage provider factory” phức tạp hơn.
