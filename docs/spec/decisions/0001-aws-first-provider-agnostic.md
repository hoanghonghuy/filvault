# ADR 0001 — AWS-first integration, provider-agnostic core

- Trạng thái: **Accepted**
- Ngày: 2026-08-14
- Spec chi tiết: [../02-provider-strategy.md](../02-provider-strategy.md)

## Bối cảnh

Filvault có hai mục tiêu: làm sản phẩm cloud storage, và học AWS thật (S3, RDS, DynamoDB, SQS, Lambda, …).

Nếu domain gọi AWS SDK trực tiếp, project trở thành bài tập AWS: khó chạy local, khó đổi RDS → Postgres thường, khó thêm DynamoDB sau này mà không xé service.

Nếu trừu tượng “hỗ trợ mọi cloud” quá sớm, mất đúng thứ AWS làm tốt (presigned URL, IAM, S3 event) và phạm YAGNI (không làm thứ chưa cần).

RDS PostgreSQL không phải Always Free dài hạn. DynamoDB có hạn mức Always Free. Muốn **hỗ trợ được** cả hai về sau, nhưng không implement DynamoDB ngay.

## Quyết định

1. Core (handler/service/domain) **provider-agnostic**: chỉ nói qua port.
2. Port **AWS-first**: hình dạng cổng ưu tiên khả năng AWS; adapter khác phải theo, không ngược lại.
3. RDS PostgreSQL = cùng `postgres` adapter với Docker Postgres; khác nhau ở deploy.
4. DynamoDB = adapter metadata **tùy chọn**, thiết kế mapping trong spec, **không code** cho đến khi có phase/học tập chủ động.
5. Phase 1 chỉ implement: `postgres` + `s3` (AWS và MinIO cùng implementation). Queue, Lambda, Cognito, CloudFront: cổng hoặc ghi nhận, chưa code.
6. Không ép Clean Architecture nhiều lớp. Giữ `handler → service → port → adapter`. Wiring ở `cmd/api`.

## Hệ quả

- Học AWS vẫn xảy ra ở adapter + Terraform + IAM, không ở domain.
- Đổi local MinIO ↔ S3 production không đụng service.
- Bật RDS = đổi connection/Terraform, không viết repository mới.
- Thêm DynamoDB sau này là thêm package adapter, không viết lại use case — với điều kiện use case không vượt quá năng lực đã ghi ở §4 spec 02.
- Chi phí: phải viết interface ngay, dù Phase 1 chỉ có một implementation thật sự cho mỗi cổng. Chấp nhận vì đúng mục tiêu “thiết kế sẵn, chưa implement hết”.
- Rủi ro: port quá chung → AWS vụng; port quá giống SDK → giả trừu tượng. Spec 02 lấy upload presigned URL làm chuẩn để tránh hai cực này.
