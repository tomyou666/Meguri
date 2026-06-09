---
name: test-overview-style
description: Write brief test overviews when adding or editing tests. Use when working on test files.
---

# Test Overview Style

テスト作成・変更時は、何を・どの条件で検証するかの概要を書く。

1. **スイート**（`describe` / `TestXxx`）: グループの意図（1 文）
2. **ケース**（`it` / `t.Run` など）: シナリオ名は必須。テストの概要はここに書く
3. 非自明なセットアップ（モック・フェイク等）があれば直前コメントで前提を書く
