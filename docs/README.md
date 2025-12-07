# ドキュメント

このディレクトリには、TechBook Compass BFF APIプロジェクトの設計書、実装計画、完了レポートなどのドキュメントが含まれています。

## ディレクトリ構成

```
docs/
├── ImplementationPlan/     # 実装計画書
│   ├── bff-api-clean-architecture.md
│   ├── book-detail-api-implementation-plan.md
│   └── ranking-api-implementation-plan.md
└── Walkthrough/            # 実装完了レポート
    ├── bff-api-implementation-report.md
    ├── book-detail-api-walkthrough.md
    ├── ranking-api-walkthrough.md
    └── daily-batch-walkthrough.md
```

## ドキュメント一覧

### 実装計画 (ImplementationPlan)

実装前の設計書や計画書を格納します。

- [**bff-api-clean-architecture.md**](./ImplementationPlan/bff-api-clean-architecture.md)
  - クリーンアーキテクチャに基づくBFF API実装計画
  - ディレクトリ構造、各層の責務、検証計画を記載

### 実装完了レポート (Walkthrough)

実装完了後の検証結果や使用方法を記載したレポートを格納します。

- [**bff-api-implementation-report.md**](./Walkthrough/bff-api-implementation-report.md)
  - BFF API実装完了レポート
  - 実装内容、検証結果、使用方法、今後の拡張予定を記載

- [**book-detail-api-walkthrough.md**](./Walkthrough/book-detail-api-walkthrough.md)
  - 書籍詳細API実装レポート

- [**ranking-api-walkthrough.md**](./Walkthrough/ranking-api-walkthrough.md)
  - ランキングAPI実装レポート

- [**daily-batch-walkthrough.md**](./Walkthrough/daily-batch-walkthrough.md)
  - 日次バッチ処理実装レポート
  - Qiita記事収集、書籍情報取得、スコアリングの処理フロー
  - コマンドラインオプション、環境変数設定、Slack通知

## 関連ドキュメント

- [API仕様 (OpenAPI)](../api/openapi.yaml) - RESTful APIの詳細仕様
- [API仕様ガイド](../api/README.md) - OpenAPI定義の使い方
- [Makefile](../Makefile) - 開発用便利コマンド集
