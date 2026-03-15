# GCP GPU + vLLM デプロイ ナレッジベース

**作成日**: 2026-03-16
**対象リポジトリ**: BenevolentDirector, milaos-realtime-avatar-spec
**目的**: GCP上でローカルLLMをデプロイする際の知見を共有し、同じ失敗を繰り返さない

---

## 1. vLLM バージョンとモデル互換性

### 教訓: モデルリリース日とvLLMバージョンを必ず照合する

| 事象 | vLLM v0.6.6 で Qwen3.5-9B をロードしたら `KeyError: 'qwen3_5'` でクラッシュ |
|------|---------------------------------------------------------------------|
| 原因 | Qwen3.5 は 2026年2月リリース。vLLM v0.6.6 は 2024年12月時点のリリースで Qwen3.5 未サポート |
| 対策 | vLLM v0.17.1（2026-03-11）で解決 |

**ルール**: モデルの HuggingFace ページで「Supported Frameworks」を確認し、vLLM の Supported Models ページ（https://docs.vllm.ai/en/latest/models/supported_models/）でアーキテクチャ名がリストされているか検証する。

**vLLM バージョン pin のベストプラクティス**:
- `:latest` は使わない（supply-chain risk）
- pin する際はモデルの `config.json` の `model_type` が vLLM の該当バージョンでサポートされているか確認
- vLLM の GitHub Releases ページで、対象モデルの追加 PR を確認

### milaos への適用
milaos-realtime-avatar-spec の `k8s/vllm/` でも同様の構成を使う場合、**デプロイ前に vLLM バージョンとモデルの互換性を確認する**こと。

---

## 2. GPU サイジング (VRAM)

### 教訓: ADR のVRAM見積もりは実測値で検証する

| 事象 | ADR-0011 で Qwen3.5-9B を ~5GB VRAM と見積もったが、実際は 17.66GB |
|------|-------------------------------------------------------------|
| 原因 | ADR は量子化前提の数値だったが、実デプロイは fp16 だった |
| 結果 | L4 (24GB) で KV cache が 0.24GB しか確保できず実用不可 |

**VRAM 計算式（目安）**:
```
fp16: パラメータ数 × 2 bytes
  例: 9B × 2 = 18GB
  例: 35B × 2 = 70GB

GPTQ Int4: パラメータ数 × 0.5 bytes + overhead (~10%)
  例: 35B × 0.5 × 1.1 = ~20GB

MoE (活性パラメータ): 全パラメータ × 2 bytes (全重みがロードされる)
  例: GLM-4.7-Flash 30B MoE → ~60GB fp16

必要 VRAM = モデルサイズ + KV cache + overhead
  KV cache ≥ 2GB (実用最小限)
  推奨: モデルサイズの 1.3-1.5 倍
```

**GPU選定マトリクス**:
| GPU | VRAM | 最大fp16モデル | 最大GPTQ 4bitモデル | GCP マシンタイプ |
|-----|------|-------------|-------------------|----------------|
| L4 ×1 | 24GB | ~9B | ~35B | g2-standard-4/8 |
| L4 ×2 | 48GB | ~20B | ~70B | g2-standard-24 |
| L4 ×4 | 96GB | ~40B | ~140B | g2-standard-48 |
| A100 40GB | 40GB | ~16B | ~60B | a2-highgpu-1g |
| A100 80GB | 80GB | ~35B | ~120B | a2-ultragpu-1g |

### milaos への適用
milaos で音声合成モデル（CosyVoice2等）やSTTモデルをGPUで動かす場合、同様にVRAM実測値を確認してからGPUサイズを決定する。

---

## 3. GKE Private Cluster の落とし穴

### 3.1 Cloud NAT 必須

| 事象 | vLLM Pod が HuggingFace からモデルをダウンロードできない (ConnectTimeout) |
|------|----------------------------------------------------------------------|
| 原因 | Private nodes はインターネットアクセスがない。Cloud NAT が未デプロイだった |
| 対策 | networking module に Cloud Router + Cloud NAT を追加して apply |

**ルール**: Private cluster を使う場合、**Cloud NAT は必須**。外部レジストリ（Docker Hub, HuggingFace, PyPI）へのアクセスが全てブロックされる。

### 3.2 Master Authorized Networks

| 事象 | ローカルから `kubectl` が接続できない |
|------|-------------------------------------|
| 原因 | `master_authorized_cidr_blocks` が VPC 内部 (10.0.0.0/8) のみ |
| 対策 | 開発者のIPアドレスを追加 |

**ルール**: dev環境では開発者IPを明示的に追加する。本番では Cloud Shell or VPN 経由。

### 3.3 GPU Node Taint と System Pods

| 事象 | kube-dns が Pending のまま起動しない |
|------|-------------------------------------|
| 原因 | GPU ノードに `nvidia.com/gpu: NoSchedule` taint があり、kube-dns がスケジュールできない |
| 対策 | non-GPU の default pool (e2-small) を追加 |

**ルール**: GPU専用クラスタでも、**system pods 用の non-GPU プールが必要**。最小構成: e2-small ×1。

### 3.4 kubectl port-forward / logs が動かない

| 事象 | `kubectl logs` や `kubectl port-forward` が "No agent available" でエラー |
|------|-------------------------------------------------------------------|
| 原因 | Private cluster では kubelet への直接接続がブロックされる |
| 対策 | Cloud Logging でログ確認、クラスタ内 Job で疎通テスト |

**代替手段**:
```bash
# ログ取得
gcloud logging read 'resource.labels.pod_name:"<pod-name>"' --project=<project> --limit=20

# クラスタ内テスト (Job で curl)
kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: test
  namespace: llm-serving
spec:
  template:
    spec:
      restartPolicy: Never
      tolerations:
      - key: nvidia.com/gpu
        operator: Exists
        effect: NoSchedule
      containers:
      - name: curl
        image: curlimages/curl:8.12.1
        command: ["curl", "-s", "http://<service>:8000/v1/models"]
EOF
```

### milaos への適用
milaos-realtime-avatar-spec も asia-northeast1 で GKE を使用。同じ private cluster 構成なら上記全てが該当する。**特に Cloud NAT は初日で設定すること。**

---

## 4. PVC とモデルストレージ

### 教訓: 複数モデルの合計サイズ + 50% 余裕を確保

| 事象 | PVC 30Gi に3モデル (合計 ~98GB) をダウンロードしようとして容量不足 |
|------|-----------------------------------------------------------|
| 対策 | PVC を 150Gi に拡張。Node disk も 200GB に |

**計算式**:
```
PVC サイズ = Σ(モデルサイズ) × 1.5 (HF cache の symlink + 一時ファイル)
Node disk = PVC + コンテナイメージ (~10GB) + OS (~5GB) + 余裕
```

**PVC 拡張方法（GKE）**:
```bash
# 1. kubectl patch (online resize)
kubectl -n <ns> patch pvc <name> -p '{"spec":{"resources":{"requests":{"storage":"150Gi"}}}}'

# 2. Pod 再マウントでファイルシステムが自動拡張
# → Pod を再起動するか、新しい Pod がマウントすると反映
```

**RWO 制約**: ReadWriteOnce PVC は同時に1つの Pod しかマウントできない。モデルダウンロード Job と vLLM Pod は同時に動かせない。**vLLM を scale=0 にしてからダウンロード**する。

---

## 5. vLLM 設定のベストプラクティス

### 5.1 Qwen3.5 の Thinking Mode

| 事象 | Qwen3.5 がデフォルトで "Thinking Process:" を出力し、max_tokens を消費 |
|------|----------------------------------------------------------------|
| 対策 | `--default-chat-template-kwargs '{"enable_thinking": false}'` |

**注意**: `--override-generation-config '{"thinking": false}'` は v0.17.1 では効かない。`--default-chat-template-kwargs` を使うこと。

リクエスト単位で制御する場合:
```json
{
  "chat_template_kwargs": {"enable_thinking": true}
}
```

### 5.2 Tensor Parallelism

複数 GPU でモデルを分散ロード:
```
--tensor-parallel-size 2  # L4 ×2 の場合
```
GPU リソースリクエストも合わせる:
```yaml
resources:
  requests:
    nvidia.com/gpu: 2  # tensor-parallel-size と一致させる
```

### 5.3 量子化モデルの選択

vLLM では **AWQ > GPTQ >> GGUF** の順で推奨:
| 形式 | vLLM 速度 | 品質 | 用途 |
|------|----------|------|------|
| AWQ (Marlin) | 741 tok/s | 95% | **vLLM 本番推奨** |
| GPTQ (Marlin) | 712 tok/s | 94% | AWQ がない場合 |
| GGUF | 93 tok/s | 92% | **vLLM では使わない** (llama.cpp 向け) |

### 5.4 Spot Instance の注意

| 事象 | Spot L4 GPU が確保できず Pod が永久 Pending |
|------|------------------------------------------|
| 原因 | asia-northeast1 で L4 Spot の供給不足 |
| 対策 | On-demand に切り替え |

**ルール**: dev環境では On-demand から始める。Spot はコスト最適化フェーズで検証。夜間停止 (Cloud Scheduler) で十分なコスト削減が可能。

---

## 6. チェックリスト: GCP GPU デプロイ

### デプロイ前
- [ ] モデルの `config.json` → `model_type` を確認
- [ ] vLLM Supported Models で対応バージョンを確認
- [ ] VRAM 実測値を計算（fp16 or 量子化）
- [ ] GPU マシンタイプ + VRAM が要件を満たすか確認
- [ ] PVC サイズ = 全モデル合計 × 1.5
- [ ] Cloud NAT が有効か確認
- [ ] master_authorized_cidr_blocks に開発者IP

### Terraform apply 前
- [ ] `terraform plan` の差分を確認
- [ ] GPU quota が region で利用可能か確認
- [ ] Spot vs On-demand の判断

### K8s apply 前
- [ ] non-GPU default pool があるか（kube-dns用）
- [ ] PVC namespace が deployment と一致
- [ ] GPU toleration が設定されているか
- [ ] tensor-parallel-size と nvidia.com/gpu 数が一致

### デプロイ後
- [ ] Pod 1/1 Running, restarts=0
- [ ] `/v1/models` HTTP 200
- [ ] `/v1/chat/completions` 正常レスポンス
- [ ] kube-dns が Running（DNS解決可能）
- [ ] レイテンシが実用範囲内（TTFT < 100ms）

---

## 7. コスト最適化

### 現在の構成 (BenevolentDirector dev)
| リソース | 単価 | 月額（24h稼働） | 夜間停止（12h/日） |
|---------|------|----------------|------------------|
| g2-standard-24 (L4 ×2) | ~$1.40/hr | ~$1,008 | ~$504 |
| e2-small ×3 (default pool) | ~$0.02/hr | ~$43 | $43 (常時) |
| PVC 150Gi (premium-rwo) | $0.17/GB/月 | $26 | $26 |
| Cloud NAT | ~$0.045/hr | ~$32 | ~$32 |
| **合計** | | **~$1,109** | **~$605** |

### 削減策
1. **夜間停止**: Cloud Scheduler で 22:00-08:00 + 週末停止 → ~45%削減
2. **Scale to zero**: min_node_count=0 + Pod 未使用時にスケールダウン
3. **Spot（将来）**: 供給安定したら切り替えで ~60%削減

---

## 変更履歴

| 日付 | 変更 |
|------|------|
| 2026-03-16 | 初版作成。BenevolentDirector vLLM PoC の知見を文書化 |
