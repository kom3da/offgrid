#!/usr/bin/env python3
"""カリキュラムの文書を検査する（管理者用。学習者のリポジトリには配らない）。

止めたいのは、今までに実際に起きた次の失敗。

1. リンク切れ・アンカー切れ
2. 引退させたルールの言い回しが、別のファイルに残る（課題シート11本の「Web検索は0回」）
3. 数が合わない（ステージの目安回数が、課題シートの目安の合計と違う）
4. ユニット数の食い違い

使い方: python3 .github/check-docs.py
"""

import os
import re
import sys
import unicodedata
from collections import defaultdict

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SKIP_DIRS = {".git", "node_modules", "dist", ".astro", "answers", "site"}

# 引退させたルールの言い回し。ルールを変えたら、古い言い方をここに足す。
# 「どのファイルにも出てこない」ことを確かめる（例外は allow に書く）。
RETIRED = [
    ("Web検索は0回", "検索の回数制限は撤回した（docs/04 のルール）"),
    ("検索が0回", "同上"),
    ("Web検索の上限", "同上"),
    ("オフライン教材セットアップ", "ステージの前提条件から外した（docs/20）"),
    ("オーナー", "「学習者」または「管理する人」と書く（CLAUDE.md の役割）"),
    ("GNUコマンド", "Ubuntu 26.04 の coreutils は uutils（CLAUDE.md を参照）"),
    ("週1回前提", "ペースは学習者が決める（docs/04）"),
    ("週次レビュー", "セッションを暦に結び付けない（docs/04）"),
    # 閉域は目的ではなく動機（docs/01）。測るのは「AIなしで自分でできるか」
    ("ネットワークを切", "ネットワークは切らない。測るのはAIなしで自分でできるか（docs/01・04）"),
    ("ネットワーク遮断", "同上"),
    ("閉域シミュレーション", "O8 は「引き継ぎ」になった"),
    ("閉域の準備", "O7 は「再現できるビルド」になった"),
    ("閉域での再構築", "C7 は「障害試験と runbook」になった"),
    ("閉域AIデー", "廃止した（弱いAIを使えるかは目的ではない）"),
    ("運用と閉域", "トラック名は「運用」"),
    ("オフライン開発", "同上"),
    ("ネットワークがなくても開発", "同上（docs/14 の旧い目的）"),
    ("閉域で動く", "同上"),
    ("閉域での完全再構築", "C7 は「障害試験と runbook」になった"),
    ("offline-bundle", "旧 O7 の持ち込み用ディレクトリ。もう無い"),
]
# 引退語を書いてよい場所（その言い回し自体を説明しているファイル）
ALLOW = {
    "Web検索は0回": {"docs/30-revision.md"},  # この仕組みの説明として引用している
    "週次レビュー": {"prompts/weekly-review.md"},  # ファイル名の都合で本文から参照する
}

problems = []
notes = []


def add(msg):
    problems.append(msg)


def md_files():
    for dirpath, dirnames, filenames in os.walk(ROOT):
        dirnames[:] = [d for d in dirnames if d not in SKIP_DIRS]
        for fn in sorted(filenames):
            if fn.endswith(".md") and not fn.startswith("._"):
                yield os.path.join(dirpath, fn)


def rel(path):
    return os.path.relpath(path, ROOT)


def read(path):
    with open(path, encoding="utf-8", errors="replace") as f:
        return f.read()


# ---------------------------------------------------------------- 1. リンク

LINK = re.compile(r"\[[^\]]*\]\(([^)\s]+)")
HEAD = re.compile(r"^#{1,6}\s+(.*)$", re.M)
_anchors = {}


def slug(text):
    """見出しからアンカー名を作る。記号は落とし、日本語は残す。"""
    t = re.sub(r"`([^`]+)`", r"\1", text).strip()
    out = []
    for ch in t:
        if unicodedata.category(ch)[0] in ("L", "N") or ch in "ー・":
            out.append(ch)
        else:
            out.append("-")
    return re.sub(r"-+", "-", "".join(out)).strip("-").lower()


def anchors(path):
    if path not in _anchors:
        _anchors[path] = {slug(m) for m in HEAD.findall(read(path))} if os.path.isfile(path) else set()
    return _anchors[path]


def check_links():
    for path in md_files():
        text = read(path)
        for i, line in enumerate(text.splitlines(), 1):
            for target in LINK.findall(line):
                if target.startswith(("http", "mailto:", "#/")):
                    continue
                if target.startswith("#"):
                    fpath, anc = path, target[1:]
                elif "#" in target:
                    p, anc = target.split("#", 1)
                    fpath = os.path.normpath(os.path.join(os.path.dirname(path), p))
                else:
                    fpath, anc = os.path.normpath(os.path.join(os.path.dirname(path), target)), None
                if not os.path.exists(fpath):
                    add(f"リンク切れ  {rel(path)}:{i} -> {target}")
                elif anc and fpath.endswith(".md") and anc not in anchors(fpath):
                    add(f"アンカー切れ  {rel(path)}:{i} -> {target}")
    notes.append("リンクとアンカー：検査した")


# ---------------------------------------------------------------- 2. 引退した言い回し


def site_files():
    """サイトの文面（docs から生成しないページと設定）。SKIP_DIRS の site の中で、これだけは見る。"""
    for r in ("site/src/content/index.mdx", "site/astro.config.mjs"):
        p = os.path.join(ROOT, r)
        if os.path.exists(p):
            yield p


def check_retired():
    for path in list(md_files()) + list(site_files()):
        r = rel(path)
        if r.startswith(".github/") or r == "CHANGELOG.md":
            continue  # この検査スクリプト自身と、引退したことを記録する変更履歴
        for i, line in enumerate(read(path).splitlines(), 1):
            for phrase, why in RETIRED:
                if phrase in line and r not in ALLOW.get(phrase, set()):
                    add(f"引退した言い回し  {r}:{i}  「{phrase}」  → {why}")
    notes.append(f"引退した言い回し（{len(RETIRED)}件）：検査した")


# ---------------------------------------------------------------- 3〜4. 数の整合

UNIT_RE = re.compile(r"^([FGTDOC])(\d{2})-")


def sheet_estimates():
    """課題シートのヘッダから、ステージごとの目安回数を集める。"""
    per_stage = defaultdict(lambda: [0, 0])  # stage -> [下限, 上限]（メイン枠のみ）
    units = set()
    for track in sorted(os.listdir(os.path.join(ROOT, "drills"))):
        d = os.path.join(ROOT, "drills", track)
        if not os.path.isdir(d):
            continue
        for name in sorted(os.listdir(d)):
            m = UNIT_RE.match(name)
            if not m:
                continue
            sheet = None
            for cand in ("TASKS.md", "TASKS.local.md"):
                p = os.path.join(d, name, cand)
                if os.path.isfile(p):
                    sheet = p
                    break
            if not sheet:
                continue
            units.add(m.group(1) + str(int(m.group(2))))
            head = re.search(r"^> (.+)$", read(sheet), re.M)
            if not head:
                add(f"ヘッダ行がない  {rel(sheet)}")
                continue
            st = re.search(r"Stage (\d)", head.group(1))
            me = re.search(r"目安：([^／]+)", head.group(1))
            if not st or not me:
                add(f"ヘッダに Stage か目安がない  {rel(sheet)}")
                continue
            text = me.group(1)
            nums = [int(x) for x in re.findall(r"\d+", text)]
            if not nums:
                continue
            lo, hi = (nums[-2], nums[-1]) if len(nums) >= 2 and "〜" in text else (nums[-1], nums[-1])
            if "1時間枠" in text:
                continue  # メイン枠と並行して進むので、回数には足さない
            per_stage[st.group(1)][0] += lo
            per_stage[st.group(1)][1] += hi
    return per_stage, units


def check_numbers():
    per_stage, units = sheet_estimates()

    if len(units) != 73:
        add(f"ユニット数が {len(units)} 本（73 のはず）")
    for path in ("README.md", "docs/30-revision.md", "CLAUDE.md", "site/src/content/index.mdx"):
        text = read(os.path.join(ROOT, path))
        for m in re.finditer(r"(?:全|約)?(\d+)の?ユニット", text):
            if m.group(1) != str(len(units)):
                add(f"ユニット数の記述が違う  {path}  「{m.group(0)}」 → 実際は {len(units)}")

    # PROGRESS.md のチェックボックス
    boxes = re.findall(r"^- \[[ x]\] ([FGTDOC]\d+)\b", read(os.path.join(ROOT, "PROGRESS.md")), re.M)
    if len(boxes) != len(units):
        add(f"PROGRESS.md のユニット欄が {len(boxes)} 個（課題シートは {len(units)} 本）")
    missing = units - set(boxes)
    if missing:
        add(f"PROGRESS.md に欄が無いユニット: {sorted(missing)}")

    # docs/03 のステージごとの目安が、課題シートの合計＋実技1回と合うか
    roadmap = read(os.path.join(ROOT, "docs", "03-roadmap.md"))
    stated = re.findall(r"### Stage (\d)[^\n]*\n\n- 目安：AIなしデー ([^\n]+)", roadmap)
    if len(stated) != 6:
        add(f"docs/03 のステージの目安が {len(stated)} 個しか読めない（6 のはず）")
    for stage, text in stated:
        nums = [int(x) for x in re.findall(r"(\d+)", text)]
        if not nums:
            add(f"docs/03 の Stage {stage} の目安から数を読めない: {text!r}")
            continue
        lo, hi = (nums[0], nums[1]) if len(nums) >= 2 and "〜" in text else (nums[0], nums[0])
        s_lo, s_hi = per_stage.get(stage, [0, 0])
        exam = 0 if stage == "5" else 1  # Stage 5 は卒業制作そのものが実技
        want_lo, want_hi = s_lo + exam, s_hi + exam
        if (lo, hi) != (want_lo, want_hi):
            add(
                f"docs/03 の Stage {stage} の目安が合わない  書いてある: {lo}〜{hi}回 / "
                f"課題シートの合計＋実技{exam}回: {want_lo}〜{want_hi}回"
            )
    # 全体の回数（docs/01・03・04 の「約N〜M回／セッション」）が、ステージの合計と合うか
    t_lo = sum(per_stage.get(s, [0, 0])[0] + (0 if s == "5" else 1) for s in "012345")
    t_hi = sum(per_stage.get(s, [0, 0])[1] + (0 if s == "5" else 1) for s in "012345")
    for path in ("docs/01-overview.md", "docs/03-roadmap.md", "docs/04-ai-free-day.md"):
        text = read(os.path.join(ROOT, path))
        found = re.findall(r"約(\d+)〜(\d+)(?:回|セッション)", text)
        if not found:
            add(f"{path} に全体の目安（約N〜M回）が無い")
        for lo, hi in found:
            if (int(lo), int(hi)) != (t_lo, t_hi):
                add(f"{path} の全体の目安が合わない  書いてある: 約{lo}〜{hi} / ステージの合計: {t_lo}〜{t_hi}")
    notes.append(f"数の整合：ユニット {len(units)} 本、ステージ {len(stated)} 個、全体 {t_lo}〜{t_hi} 回を検査した")


# ----------------------------------------------------------------


def check_agents():
    """AGENTS.md（Claude 以外のツールが読む指示）が、要点を落としていないか。

    CLAUDE.md だけ直して AGENTS.md が古くなる、という取り残しを機械で止める。
    ここで見るのは「最重要ルール：学習を奪わない」の写しだけ。
    """
    path = os.path.join(ROOT, "AGENTS.md")
    if not os.path.exists(path):
        add("AGENTS.md がない（Claude 以外のAIがルールを読めなくなる）")
        return
    text = read(path)
    for phrase, why in (
        ("頼まれずに書き換えない", "学習者のコードを勝手に直さない"),
        ("`answers/` は読まない", "答えを引用させない"),
        ("解かない", "課題シートとデバッグドリルを解かせない"),
        ("pgx", "外部ライブラリは G17 の pgx だけ"),
        ("勧めない", "答えの載ったページを勧めさせない"),
        ("CLAUDE.md", "残りの決まりへの案内"),
    ):
        if phrase not in text:
            add(f"AGENTS.md に「{phrase}」が無い（{why}）")
    notes.append("AGENTS.md：要点6つを検査した")


def main():
    check_links()
    check_retired()
    check_numbers()
    check_agents()
    for n in notes:
        print(f"  {n}")
    if problems:
        print(f"\n問題 {len(problems)} 件:", file=sys.stderr)
        for p in problems:
            print(f"  - {p}", file=sys.stderr)
        return 1
    print("\ncheck-docs: 問題ありません")
    return 0


if __name__ == "__main__":
    sys.exit(main())
