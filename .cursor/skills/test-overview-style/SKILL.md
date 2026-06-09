---
name: test-overview-style
description: Write brief test overviews when adding or editing tests. Use when working on test files.
---

# Test Overview Style

テスト作成・変更時は、何を・どの条件で検証するかの概要を書く。

1. **ファイル先頭**: 対象と検証範囲（1〜2 文）
2. **スイート**（`describe` / `TestXxx`）: グループの意図（1 文）
3. **ケース**（`it` / `t.Run`）: シナリオ名で意図を明示
4. 非自明なセットアップ（モック・フェイク等）があれば直前コメントで前提を書く
