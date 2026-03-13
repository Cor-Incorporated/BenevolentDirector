# dev

開発用 GCP 環境の Terraform 変数と state 配置をここに置く。
Secret Manager の secret value は Terraform に含めず、既存 secret は `terraform import` で state に取り込む。
