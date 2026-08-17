Điểm quan trọng nhất: **không xây một “Google Drive clone” đầy đủ ngay lập tức**. Phase hiện tại chỉ làm một lõi lưu trữ file tốt, có AWS thật, có kiến trúc đủ sạch để sau này thêm sync, sharing, versioning, worker, CLI… mà không phải đập đi làm lại.

---

# Filvault — Product & Technical Concept Specification

## 1. Tổng quan ý tưởng

**Filvault** là một nền tảng lưu trữ và đồng bộ file cá nhân trên cloud.

Ở mức cơ bản, người dùng có thể:

```text
Upload file
Download file
Create folders
Move / rename files
Delete / restore files
Manage storage
```

Khi hoàn thiện, Filvault hướng tới:

```text
                 ┌─────────────────┐
                 │      Web        │
                 │      Vue        │
                 └────────┬────────┘
                          │
           ┌──────────────┼──────────────┐
           │              │              │
           ▼              ▼              ▼
         CLI          Desktop Sync     Mobile*
           │              │
           └──────────────┼──────────────┘
                          │
                          ▼
                     Filvault API
                          │
               ┌──────────┴─────────┐
               │                    │
               ▼                    ▼
          PostgreSQL           Object Storage
                                  │
                              AWS S3
```

`*` Mobile là khả năng mở rộng tương lai, không phải mục tiêu hiện tại.

---



# 2. Mục tiêu của dự án

Dự án có hai mục tiêu song song.

### Mục tiêu sản phẩm

Xây dựng một hệ thống cloud storage có trải nghiệm gần với:

- Google Drive
- Dropbox
- OneDrive

nhưng nhỏ gọn hơn và tập trung vào những chức năng cốt lõi.

### Mục tiêu học tập

Project phải giúp thực hành thực tế:

```text
Vue
Go
PostgreSQL

REST API
Authentication
Authorization
Database design

AWS IAM
AWS S3
AWS Lambda
AWS SQS
AWS CloudFront
AWS RDS
AWS CloudWatch

Docker
Terraform
GitHub Actions
OIDC
```

Vì vậy không thêm AWS service chỉ để CV có nhiều logo.

Mỗi service phải giải quyết một vấn đề thực tế.

---



# 3. Những quyết định đã chốt



## 3.1 Tên

Working name:

```text
Filvault
```

CLI:

```bash
filvault
```

Ví dụ:

```bash
filvault login
filvault ls
filvault upload report.pdf
filvault download report.pdf
filvault sync ~/Documents
```

Tên vẫn có thể thay đổi trước khi public chính thức.

---



# 4. Technology stack



## Frontend

```text
Vue 3
TypeScript
Vite
Vue Router
Pinia
```

Có thể dùng UI framework sau, nhưng chưa cần khóa ngay:

```text
Tailwind CSS
shadcn-vue
```

---



## Backend

```text
Go
Gin
```

Backend ban đầu là:

> **Modular Monolith**

Không microservices.

---



## Database

```text
PostgreSQL
```

PostgreSQL là database chính/source of truth.

Nó lưu:

```text
users
folders
files
file_versions
shares
devices
upload_sessions
audit_logs
...
```

---



## Object Storage

Production:

```text
AWS S3
```

Development local:

```text
MinIO
```

File binary **không lưu vào PostgreSQL**.

Ví dụ:

```text
PostgreSQL
---------------------------------
id
owner_id
name
mime_type
size
object_key
status
created_at
...

            │
            │ object_key
            ▼

AWS S3
---------------------------------
users/{userId}/files/{fileId}
```

---



# 5. Kiến trúc lõi

Kiến trúc ban đầu:

```text
                    Vue Web
                       │
                       │ HTTPS / REST
                       ▼
                    Go API
                 ┌─────┴──────┐
                 │            │
                 ▼            ▼
            PostgreSQL       AWS S3
              metadata       objects
```

Khi phát triển đầy đủ hơn:

```text
                         CloudFront
                             │
                             ▼
                           Vue
                             │
                             ▼
                      API Gateway
                             │
                             ▼
                          Go API
                       /          \
                      /            \
                     ▼              ▼
             PostgreSQL            S3
               RDS                 │
                                   │ Event
                                   ▼
                                  SQS
                                   │
                                   ▼
                              Go Worker
                               Lambda
                                   │
                        ┌──────────┴─────────┐
                        ▼                    ▼
                   thumbnails           metadata


                        CloudWatch
                             ↑
                   Logs / Metrics / Alerts
```

---



# 6. Nguyên tắc thiết kế quan trọng



## PostgreSQL quản lý metadata

Ví dụ:

```text
report.pdf
```

PostgreSQL:

```text
id          019abc...
name        report.pdf
size        4.7 MB
mime_type   application/pdf
object_key  users/01.../files/019abc
```

S3 chỉ có:

```text
users/01.../files/019abc
```

Không nên dùng:

```text
users/huy/Documents/report.pdf
```

làm S3 key.

---



## Tên file không quyết định object key

Nếu:

```text
report.pdf
```

được rename:

```text
graduation-report.pdf
```

chỉ cần update PostgreSQL.

Không cần:

```text
copy object
delete old object
```

trong S3.

---



# 7. Upload architecture

Filvault sẽ dùng **S3 Presigned URL**.

Không upload file lớn xuyên qua Go API:

```text
Vue
 │
 │ file
 ▼
Go
 │
 ▼
S3
```

Mà:

```text
1.

Vue
 │
 │ request upload
 ▼
Go
 │
 ▼
PostgreSQL

2.

Go
 │
 │ create presigned URL
 ▼
AWS S3

3.

Vue ────────────────> AWS S3
        upload
```

Ví dụ:

```http
POST /api/v1/files/upload-sessions
```

Request:

```json
{
  "name": "photo.jpg",
  "size": 2849512,
  "contentType": "image/jpeg",
  "folderId": "..."
}
```

Response:

```json
{
  "fileId": "...",
  "uploadUrl": "https://...",
  "expiresAt": "..."
}
```

Sau đó Vue:

```http
PUT {uploadUrl}
```

trực tiếp tới S3.

---



# 8. File state

Không nên coi file là tồn tại ngay khi tạo record.

Có thể dùng state:

```text
PENDING
   │
   │ upload thành công
   ▼
UPLOADED
   │
   │ processing
   ▼
PROCESSING
   │
   ▼
READY
```

Nếu lỗi:

```text
FAILED
```

Ví dụ enum:

```text
PENDING
UPLOADED
PROCESSING
READY
FAILED
```

Phase đầu chưa nhất thiết cần toàn bộ.

Có thể bắt đầu:

```text
PENDING
READY
FAILED
```

---



# 9. Phạm vi Phase hiện tại

Mình đề xuất gọi:

# Phase 1 — Core Storage MVP

Đây là phase mà bạn nên bắt đầu xây ngay.

Mục tiêu:

> Người dùng đăng nhập được, quản lý folder/file và thực sự upload/download file qua AWS S3.

---



# 10. Phase 1 — chức năng bắt buộc



## 10.1 Authentication

Người dùng có thể:

```text
Register
Login
Logout
Refresh token
Get current user
```

Ban đầu có thể tự implement:

```text
Email + Password
JWT Access Token
Refresh Token
```

Chưa cần Cognito.

Lý do:

Bạn đang học backend, nên tự xây auth giúp hiểu:

```text
password hashing
JWT
refresh token
authentication
authorization
```

---



# 11. User

Thông tin cơ bản:

```text
id
email
display_name
password_hash
storage_used
storage_quota
created_at
updated_at
```

Người dùng có:

```text
My Files
Storage usage
Profile
```

---



# 12. Folder

User có thể:

```text
Create folder
Rename folder
Open folder
Move folder
Delete folder
Restore folder
```

Folder hỗ trợ nested structure:

```text
My Files
│
├── Documents
│   ├── University
│   │   └── Thesis
│   │
│   └── Personal
│
├── Photos
│
└── Projects
```

Database:

```text
folders

id
owner_id
parent_id
name
created_at
updated_at
deleted_at
```

`parent_id = NULL`:

```text
root
```

---



# 13. File

Phase 1 cần:

```text
Upload
Download
Rename
Move
Delete
Restore
View metadata
```

File:

```text
files

id
owner_id
folder_id

name
original_name

object_key

mime_type
size_bytes

status

created_at
updated_at
deleted_at
```

---



# 14. File upload

User:

```text
Choose file
    ↓
Upload
    ↓
Progress
    ↓
Complete
```

Frontend phải hiển thị:

```text
Uploading report.pdf

██████████████░░░░░ 72%

3.4 MB / 4.7 MB
```

Có:

```text
upload progress
upload failure
retry
cancel
```

Retry nâng cao có thể để phase sau.

---



# 15. Download

Không public S3 bucket.

Backend xác thực:

```text
User
 ↓
GET /files/{id}/download
 ↓
Go
 ↓
check owner
 ↓
create presigned download URL
```

Sau đó:

```text
Vue → S3
```

---



# 16. File browser

Web UI cần ít nhất:

```text
My Files

Name               Size        Modified
------------------------------------------------
📁 Documents
📁 Photos
📄 report.pdf      4.7 MB      today
🖼 image.png        2.1 MB      yesterday
```

Có:

```text
List view
Breadcrumb
Folder navigation
```

Ví dụ:

```text
My Files > Documents > University > Thesis
```

---



# 17. Basic search

Phase 1 có thể hỗ trợ search đơn giản:

```text
Search file/folder by name
```

Thông qua PostgreSQL.

Ví dụ:

```http
GET /api/v1/search?q=report
```

Chưa cần Elasticsearch/OpenSearch.

---



# 18. Trash

Soft delete:

```text
File
 │
 │ delete
 ▼
Trash
```

Không xóa object S3 ngay.

Có:

```text
Restore
Delete permanently
```

Ví dụ:

```text
deleted_at
```

được set khi đưa vào Trash.

Permanent delete mới:

```text
DELETE S3 object
DELETE / archive DB record
```

---



# 19. Storage quota

Ví dụ:

```text
Storage

4.7 GB / 10 GB

██████████░░░░░░░░
```

Backend phải kiểm tra quota trước khi cho upload.

Không được chỉ tin:

```json
{
  "size": 123
}
```

từ client.

Sau khi upload có thể verify object metadata từ S3.

---



# 20. Phase 1 AWS services

Phase này **không cần dùng mọi AWS service**.

Chỉ cần:

### Bắt buộc

```text
IAM
S3
```



### Có thể dùng khi deploy

```text
RDS PostgreSQL
CloudWatch
```

Chưa cần:

```text
Lambda
SQS
CloudFront
API Gateway
DynamoDB
ECS
```

---



# 21. Development environment

Local:

```text
Vue
 │
 ▼
Go API
 ├── PostgreSQL
 └── MinIO
```

Docker Compose:

```text
PostgreSQL
MinIO
```

Có thể thêm:

```text
Adminer / pgAdmin
```

nếu muốn.

---



# 22. Production

Sau Phase 1:

```text
Vue
  │
  ▼
Go API
  │
  ├──────────────► Amazon RDS PostgreSQL
  │
  └──────────────► Amazon S3
```

Cách deploy Go sẽ quyết định ở phase deployment.

Có thể là:

```text
Lambda
```

hoặc:

```text
ECS
```

Không cần khóa lựa chọn này ngay hôm nay.

---



# 23. Những thứ KHÔNG nằm trong Phase 1

Đây là phần rất quan trọng để chống scope creep.

Không làm ngay:

```text
❌ Desktop sync
❌ CLI sync
❌ File version history
❌ Public sharing links
❌ Team workspace
❌ Collaborative editing
❌ SQS
❌ Lambda workers
❌ Thumbnail pipeline
❌ Virus scanning
❌ OCR
❌ Full-text search
❌ CloudFront CDN
❌ Multi-region
❌ Multi-storage providers
❌ WebSocket real-time sync
❌ Mobile application
❌ Microservices
```

---



# 24. Definition of Done của Phase 1

Phase 1 hoàn thành khi có thể thực hiện toàn bộ flow:

```text
Register
   ↓
Login
   ↓
Create folder
   ↓
Upload file
   ↓
File goes to S3
   ↓
Metadata goes to PostgreSQL
   ↓
Browse file
   ↓
Rename
   ↓
Move
   ↓
Download
   ↓
Delete
   ↓
Trash
   ↓
Restore
```

Và phải đảm bảo:

```text
User A không thấy file User B
User A không download file User B
S3 bucket private
Upload sử dụng presigned URL
Password được hash
JWT authentication hoạt động
Database migration có thể chạy lại
Local environment reproducible bằng Docker
```

Nếu đạt những điểm đó:

> **Phase 1 hoàn thành.**

---



# 25. Phase 2 — Async File Processing

Sau khi MVP ổn định mới thêm:

```text
S3
 │
 │ event
 ▼
SQS
 │
 ▼
Lambda Go
```

Worker xử lý:

```text
image metadata
thumbnail generation
file validation
```

Ví dụ:

```text
photo.jpg
   │
   ▼
S3
   │
   ▼
SQS
   │
   ▼
Lambda
   │
   ├── 256x256 thumbnail
   └── 1024x1024 preview
```

---



# 26. Phase 3 — Sharing

Thêm khả năng:

```text
Share with user
Share by link
```

Ví dụ:

```text
https://filvault.../s/Jk8xP2
```

Có:

```text
expiration
password protection
download permission
```

Permissions:

```text
VIEW
DOWNLOAD
EDIT
```

---



# 27. Phase 4 — File Versioning

Ví dụ:

```text
report.docx

v1
v2
v3
v4
```

Người dùng có thể:

```text
View versions
Download previous version
Restore version
```

Data model:

```text
files
   │
   ▼
file_versions
```

Một logical file:

```text
file
```

có nhiều physical objects:

```text
file_versions
```

---



# 28. Phase 5 — CLI

Đây là lúc binary:

```bash
filvault
```

thực sự có giá trị.

Ví dụ:

```bash
filvault login
```

```bash
filvault whoami
```

```bash
filvault ls
```

```bash
filvault cd Documents
```

```bash
filvault upload report.pdf
```

```bash
filvault download report.pdf
```

```bash
filvault rm old.pdf
```

```bash
filvault mkdir projects
```

Có thể hỗ trợ:

```bash
filvault upload *.pdf
```

---



# 29. Phase 6 — Synchronization

Đây sẽ là một trong những phần khó nhất và thú vị nhất.

CLI/Desktop daemon:

```text
Local folder
     │
     │ watcher
     ▼
Filvault client
     │
     ▼
Filvault API
     │
     ▼
S3
```

Ví dụ:

```bash
filvault sync ~/Documents
```

Sau đó:

```text
~/Documents/report.pdf
```

thay đổi:

```text
Filesystem Watcher
      ↓
Detect modification
      ↓
Calculate hash
      ↓
Upload new version
```

---



# 30. Sync cần quản lý device

Database:

```text
devices

id
user_id
name
platform
last_seen_at
created_at
```

Ví dụ:

```text
Huy-Nitro
Linux
```

và:

```text
Work-PC
Windows
```

---



# 31. Conflict handling

Sau này phải xử lý trường hợp:

```text
Laptop
report.docx
      │
      │ edit
      ▼

Cloud
report.docx

Desktop
report.docx
      │
      │ edit simultaneously
      ▼
```

Có thể tạo:

```text
report.docx
report (conflict Huy-Nitro 2026-08-14).docx
```

Không giải quyết vấn đề này trong MVP.

---



# 32. Phase 7 — Search nâng cao

Ban đầu:

```text
PostgreSQL filename search
```

Sau này:

```text
filename
extension
MIME
date
owner
folder
tags
```

Xa hơn có thể:

```text
PDF text
DOCX text
OCR image
```

và dùng:

```text
PostgreSQL FTS
```

trước khi nghĩ tới OpenSearch.

---



# 33. Phase 8 — Preview

Browser xem:

```text
Images
PDF
Text
Video
Audio
```

Không phải lúc nào cũng cần download.

Pipeline:

```text
original
   │
   ▼
worker
   │
   ├── thumbnail
   └── preview
```

---



# 34. Phase 9 — Audit & Activity

Filvault hoàn thiện nên có:

```text
Uploaded file
Downloaded file
Renamed file
Moved file
Deleted file
Restored file
Shared file
Logged in
Device connected
```

Ví dụ:

```text
Activity

15:02  report.pdf uploaded
14:55  Photos folder created
14:48  thesis.docx renamed
```

---



# 35. Audit log

Database:

```text
audit_logs

id
actor_id

action
resource_type
resource_id

metadata

ip_address
user_agent

created_at
```

Ví dụ:

```text
FILE_DOWNLOADED
FILE_SHARED
FILE_DELETED
LOGIN_SUCCEEDED
```

---



# 36. Security

Sản phẩm hoàn chỉnh bắt buộc quan tâm đến:

### Authentication

```text
password hashing
JWT
refresh token rotation
session/device management
```



### Authorization

Mỗi resource đều kiểm tra:

```text
Who owns this file?
Who can access it?
What permission do they have?
```

Không bao giờ chỉ dựa vào:

```text
fileId
```

---



# 37. S3 security

Bucket:

```text
PRIVATE
```

Không:

```text
public-read
```

Mọi truy cập thông qua:

```text
presigned URL
```

thời gian ngắn.

Ví dụ:

```text
5–15 minutes
```

---



# 38. IAM

Backend chỉ được cấp permission cần thiết.

Ví dụ:

```text
GetObject
PutObject
DeleteObject
HeadObject
```

trên bucket Filvault.

Không dùng:

```text
AdministratorAccess
```

cho application.

---



# 39. Encryption

Production:

```text
HTTPS
```

S3:

```text
server-side encryption
```

Database credential:

```text
environment / secret manager
```

Không commit:

```text
.env
AWS keys
database password
JWT secret
```

---



# 40. File integrity

Về sau mỗi file/version có:

```text
checksum
```

Ví dụ:

```text
SHA-256
```

Dùng để:

```text
verify upload
detect corruption
sync comparison
potential deduplication
```

---



# 41. Duplicate detection

Future feature:

User upload:

```text
ubuntu.iso
```

hai lần.

Hash:

```text
SHA256 AABBCC...
```

Có thể phát hiện duplicate.

Nhưng physical deduplication cần thiết kế security/ref-count cẩn thận, nên chưa làm trong MVP.

---



# 42. Product hoàn thiện cần những module nào?

Mình chia thành 12 module chính.


| Module             | Final product |
| ------------------ | ------------- |
| Identity           | ✅             |
| Files              | ✅             |
| Folders            | ✅             |
| Upload/Download    | ✅             |
| Trash              | ✅             |
| Search             | ✅             |
| Sharing            | ✅             |
| Versioning         | ✅             |
| Synchronization    | ✅             |
| Devices            | ✅             |
| Activity/Audit     | ✅             |
| Storage management | ✅             |


---



# 43. Các tính năng sản phẩm hoàn chỉnh



## Core File Management

```text
Upload
Download
Rename
Copy
Move
Delete
Restore
Permanent delete
Multi-select
Drag & drop
```

---



## Folder Management

```text
Create
Rename
Move
Delete
Nested folder
Breadcrumb navigation
```

---



## Upload

```text
Single upload
Multiple upload
Drag & drop
Upload progress
Cancel
Retry
Large file upload
```

Sau này:

```text
multipart upload
resumable upload
```

đặc biệt cho file lớn.

---



# 44. Sharing

```text
Share with another user
Share via URL
Expiration date
Password protection
Revoke link
Read-only
Download allowed
```

---



# 45. File Versioning

```text
Version history
Restore previous version
Download previous version
Version metadata
```

---



# 46. Trash

```text
Soft delete
Restore
Permanent delete
Auto cleanup after N days
```

---



# 47. Search

```text
Name
Extension
File type
Date
Size
Folder
Owner
```

Sau này:

```text
Full-text document search
```

---



# 48. File preview

```text
Image
PDF
Text
Audio
Video
```

---



# 49. Favorites

User có thể:

```text
⭐ favorite
```

và vào:

```text
Starred
```

---



# 50. Recent files

Có:

```text
Recent
```

dựa trên:

```text
uploaded_at
modified_at
accessed_at
```

---



# 51. Storage dashboard

Ví dụ:

```text
Storage
━━━━━━━━━━━━━━━━━━━━━━━

6.8 GB / 20 GB

Images       2.3 GB
Videos       3.1 GB
Documents    0.9 GB
Others       0.5 GB
```

---



# 52. Device management

Ví dụ:

```text
Devices

Huy-Nitro
Fedora Linux
Last active: now

Work-PC
Windows 11
Last active: yesterday
```

User có thể:

```text
Revoke device
```

---



# 53. Sync client

Cuối cùng:

```bash
filvault sync ~/Documents
```

hoặc desktop app chạy background.

Hỗ trợ:

```text
Two-way sync
Selective sync
Ignore patterns
Conflict detection
Resume
Offline queue
```

---



# 54. CLI hoàn chỉnh

CLI có thể trở thành một project nhỏ riêng trong hệ sinh thái Filvault.

Ví dụ:

```bash
filvault auth login
filvault auth logout

filvault ls
filvault pwd

filvault mkdir photos
filvault rm old.txt
filvault mv a.txt documents/

filvault upload photo.jpg
filvault download report.pdf

filvault share report.pdf

filvault sync ~/Documents

filvault status

filvault storage
```

---



# 55. Infrastructure as Code

Sau khi infrastructure ổn định:

```text
Terraform
```

Repository:

```text
infrastructure/
└── terraform/
    ├── environments/
    │   ├── dev/
    │   └── prod/
    │
    └── modules/
        ├── s3/
        ├── iam/
        ├── database/
        ├── lambda/
        └── queue/
```

---



# 56. CI/CD

GitHub Actions:

```text
Push / PR
    │
    ├── frontend lint
    ├── frontend test
    ├── go test
    ├── go vet
    ├── build
    │
    ▼
Deploy
```

Authentication GitHub → AWS:

```text
GitHub Actions
      │
      │ OIDC
      ▼
AWS IAM Role
```

Không lưu AWS access key dài hạn nếu tránh được.

---



# 57. Observability

Production cần:

```text
CloudWatch Logs
CloudWatch Metrics
```

Track:

```text
API errors
latency
upload failures
worker failures
queue depth
storage usage
```

---



# 58. Background processing

Architecture cuối:

```text
Upload
   │
   ▼
S3
   │
   ▼
SQS
   │
   ▼
Worker
   │
   ├── validate
   ├── extract metadata
   ├── thumbnail
   ├── preview
   └── checksum
```

Failure:

```text
SQS
 │
 ├── retry
 │
 └── DLQ
```

---



# 59. Repository structure

Mình đề xuất monorepo:

```text
filvault/

├── apps/
│   ├── web/
│   │   ├── src/
│   │   ├── public/
│   │   └── package.json
│   │
│   ├── api/
│   │   ├── cmd/
│   │   │   └── api/
│   │   │
│   │   ├── internal/
│   │   │   ├── auth/
│   │   │   ├── user/
│   │   │   ├── file/
│   │   │   ├── folder/
│   │   │   ├── storage/
│   │   │   └── platform/
│   │   │
│   │   └── migrations/
│   │
│   └── cli/
│
├── workers/
│   └── file-processor/
│
├── infrastructure/
│   └── terraform/
│
├── docs/
│   ├── architecture/
│   ├── api/
│   └── decisions/
│
├── docker-compose.yml
├── Makefile
├── README.md
└── .gitignore
```

---



# 60. Backend architecture

Không ép Clean Architecture quá mức.

Có thể dùng:

```text
handler
   ↓
service
   ↓
repository
```

Ví dụ:

```text
internal/file/

handler.go
service.go
repository.go
model.go
dto.go
```

Storage abstraction riêng:

```go
type ObjectStorage interface {
    CreateUploadURL(...)
    CreateDownloadURL(...)
    Stat(...)
    Delete(...)
}
```

Implementations:

```text
S3Storage
MinIOStorage
```

Hoặc vì cả hai tương thích S3 API, thậm chí có thể dùng chung implementation với endpoint/config khác nhau.

---



# 61. Database dự kiến khi hoàn thiện

Core:

```text
users

folders

files

file_versions

upload_sessions

shares

share_links

devices

sync_states

favorites

audit_logs
```

Có thể hình dung:

```text
users
 │
 ├──────── folders
 │
 ├──────── files
 │             │
 │             └──── file_versions
 │
 ├──────── devices
 │
 ├──────── shares
 │
 ├──────── favorites
 │
 └──────── audit_logs
```

---



# 62. API groups

API hoàn thiện có thể tổ chức:

```text
/api/v1/auth
/api/v1/users

/api/v1/files
/api/v1/folders

/api/v1/uploads
/api/v1/downloads

/api/v1/shares
/api/v1/trash

/api/v1/search
/api/v1/storage

/api/v1/devices
/api/v1/sync
```

Không nhất thiết endpoint cuối cùng phải đúng như trên; đây là boundary của domain.

---



# 63. Roadmap tổng thể

Mình sẽ tổ chức development thành:

### Phase 1 — Core Storage MVP

```text
Vue
Go
PostgreSQL
MinIO
AWS S3
IAM
Auth
Files
Folders
Trash
Basic search
```



### Phase 2 — Cloud Deployment

```text
RDS
Backend deployment
CloudWatch
CI/CD
Terraform
```



### Phase 3 — Async Processing

```text
S3 Event
SQS
Lambda
Thumbnail
Metadata
```



### Phase 4 — Sharing & Versioning

```text
Sharing
Public links
Permissions
File versions
```



### Phase 5 — CLI

```text
filvault
```



### Phase 6 — Sync

```text
Device management
Filesystem watcher
Two-way synchronization
Conflict management
```



### Phase 7 — Advanced Product

```text
Preview
Advanced search
Activity
Audit
Storage analytics
Multipart uploads
Resumable upload
```

---



# 64. Vision sản phẩm cuối cùng

Filvault cuối cùng không chỉ là:

> một website upload file lên S3.

Mà là:

> **một personal cloud storage platform với web interface, REST API, CLI và synchronization client, sử dụng PostgreSQL quản lý trạng thái/metadata và object storage quản lý binary data, với AWS cung cấp infrastructure cho storage, asynchronous processing, deployment và observability.**

User experience cuối cùng:

```text
                     Filvault
                        │
      ┌─────────────────┼─────────────────┐
      │                 │                 │
      ▼                 ▼                 ▼
     Web               CLI             Desktop
      │                 │                 │
      └─────────────────┼─────────────────┘
                        │
                        ▼
                    Filvault API
                        │
          ┌─────────────┼──────────────┐
          │             │              │
          ▼             ▼              ▼
     PostgreSQL        S3             SQS
                                         │
                                         ▼
                                      Workers
```

User có thể:

```text
đăng nhập
    ↓
upload
    ↓
organize
    ↓
preview
    ↓
download
    ↓
share
    ↓
version
    ↓
sync giữa nhiều máy
    ↓
quản lý storage
    ↓
xem activity
```

---



## Tóm lại phạm vi chúng ta nên làm NGAY

Đừng bắt đầu từ Lambda, SQS hay sync.

Bắt đầu bằng vertical slice:

```text
Vue
 ↓
Go
 ↓
PostgreSQL
 ↓
S3
```

với đúng một flow hoàn chỉnh:

```text
Register/Login
       ↓
Create Folder
       ↓
Choose File
       ↓
Go creates upload session
       ↓
Go generates S3 Presigned URL
       ↓
Vue uploads directly to S3
       ↓
Go confirms upload
       ↓
PostgreSQL stores metadata
       ↓
File appears in My Files
       ↓
Download via Presigned URL
```

**Đó nên là mục tiêu kỹ thuật đầu tiên của Filvault.**

Khi vertical slice này chạy ổn, những phần như Trash, Move/Rename, Search được xây xung quanh nó. Sau đó mới bước sang SQS/Lambda, sharing, versioning và cuối cùng là `filvault sync`.

Bản spec này mình sẽ coi là **v0.1 — Concept & Product Scope**. Nó đủ để từ đây tách tiếp thành **Database Spec, API Spec, Architecture Spec và Development Plan** mà không làm mất định hướng tổng thể.