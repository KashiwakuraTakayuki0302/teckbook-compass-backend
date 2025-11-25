# ドキュメント

このディレクトリには、TechBook Compass BFF APIプロジェクトの設計書、実装計画、完了レポートなどのドキュメントが含まれています。

## ディレクトリ構成

```
docs/
├── ImplementationPlan/     # 実装計画書
│   └── bff-api-clean-architecture.md
└── Walkthrough/            # 実装完了レポート
    └── bff-api-implementation-report.md
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

## 関連ドキュメント

- [API仕様 (OpenAPI)](../api/openapi.yaml) - RESTful APIの詳細仕様
- [API仕様ガイド](../api/README.md) - OpenAPI定義の使い方
- [Makefile](../Makefile) - 開発用便利コマンド集
