#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = [
#   "httpx",
#   "matplotlib",
#   "packaging",
# ]
# ///

import os
import httpx
import matplotlib.pyplot as plt
import matplotlib.ticker as ticker
import sys
from pathlib import Path
from packaging import version

REPO = "rm-hull/git-commit-summary"
OUTPUT = Path(__file__).parents[2] / "docs" / "download_stats.png"


def gha(level: str, msg: str) -> None:
    if os.environ.get("GITHUB_ACTIONS"):
        print(f"::{level}::{msg}")
    else:
        print(msg)


def fetch_releases(repo: str) -> list[dict]:
    releases = []
    page = 1
    token = os.environ.get("GITHUB_TOKEN")
    headers = {"Accept": "application/vnd.github+json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    with httpx.Client() as client:
        while True:
            r = client.get(
                f"https://api.github.com/repos/{repo}/releases",
                params={"per_page": 100, "page": page},
                headers=headers,
            )
            r.raise_for_status()
            data = r.json()
            if not data:
                break
            releases.extend(data)
            if len(data) < 100:
                break
            page += 1
    return releases


def main():
    repo = sys.argv[1] if len(sys.argv) > 1 else REPO
    gha("notice", f"Fetching releases for {repo}...")

    releases = fetch_releases(repo)

    def get_arch(name: str) -> str | None:
        if name == "checksums.txt":
            return None
        parts = name.split("_")
        # Expecting: [app_name, version, os, arch.ext]
        if len(parts) < 4:
            return None
        os_part = parts[-2]
        arch_part = parts[-1].split(".")[0]
        return f"{os_part}_{arch_part}"

    stats = []
    for r in releases:
        arch_counts = {}
        total_downloads = 0
        for a in r["assets"]:
            arch = get_arch(a["name"])
            if arch:
                count = a["download_count"]
                arch_counts[arch] = arch_counts.get(arch, 0) + count
                total_downloads += count
        if total_downloads > 0:
            stats.append({
                "tag": r["tag_name"],
                "downloads": total_downloads,
                "arch_counts": arch_counts,
            })
    stats.reverse()  # chronological order

    if not stats:
        fig, ax = plt.subplots(figsize=(10, 5))
        ax.text(
            0.5,
            0.5,
            "No download data yet",
            ha="center",
            va="center",
            transform=ax.transAxes,
            fontsize=13,
            color="#888",
        )
        ax.set_title(f"Downloads per release — {repo}", fontsize=13, pad=12)
        ax.spines[["top", "right", "left", "bottom"]].set_visible(False)
        ax.set_xticks([])
        ax.set_yticks([])
        fig.tight_layout()
        fig.savefig(OUTPUT, dpi=150, bbox_inches="tight")
        gha("warning", f"  No downloads yet, saved placeholder to {OUTPUT}")
        return

    # Keep top 20 by downloads, always include the latest release
    latest = stats[-1]
    top20 = set(
        s["tag"] for s in sorted(stats, key=lambda s: s["downloads"], reverse=True)[:20]
    )
    is_latest_in_top20 = latest["tag"] in top20

    filtered = [s for s in stats if s["tag"] in top20]
    if not is_latest_in_top20:
        filtered.append(latest)
    filtered = sorted(filtered, key=lambda s: version.parse(s["tag"]))

    tags = [s["tag"] for s in filtered]
    counts = [s["downloads"] for s in filtered]
    total = sum(s["downloads"] for s in stats)
    top = max(filtered, key=lambda s: s["downloads"])

    gha("notice", f"  Total releases with downloads: {len(stats)}")
    gha("notice", f"  Showing:                       {len(filtered)} releases")
    gha("notice", f"  Total downloads:               {total:,}")
    gha(
        "notice",
        f"  Top release:                   {top['tag']} ({top['downloads']:,} downloads)",
    )

    fig, ax = plt.subplots(figsize=(max(10, len(tags) * 0.5), 5))

    all_archs = sorted({arch for s in filtered for arch in s["arch_counts"]})
    bottoms = [0] * len(filtered)
    cmap = plt.get_cmap("tab10")

    for i, arch in enumerate(all_archs):
        arch_counts = [s["arch_counts"].get(arch, 0) for s in filtered]
        ax.bar(tags, arch_counts, bottom=bottoms, label=arch, color=cmap(i), width=0.7)
        bottoms = [b + c for b, c in zip(bottoms, arch_counts)]

    ax.set_title(f"Downloads per release — {repo}", fontsize=13, pad=12)
    ax.set_xlabel("Release", fontsize=10)
    ax.set_ylabel("Downloads", fontsize=10)
    ax.yaxis.set_major_locator(ticker.MaxNLocator(integer=True))
    ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x, _: f"{int(x):,}"))
    ax.tick_params(axis="x", rotation=45, labelsize=8)
    ax.tick_params(axis="y", labelsize=9)
    ax.spines[["top", "right"]].set_visible(False)
    ax.grid(axis="y", color="#e0e0e0", linewidth=0.5)
    ax.set_axisbelow(True)

    # Annotate the top bar
    top_idx = counts.index(max(counts))
    ax.annotate(
        f"{counts[top_idx]:,}",
        xy=(top_idx, counts[top_idx]),
        xytext=(0, 4),
        textcoords="offset points",
        ha="center",
        fontsize=8,
    )

    # Legend for architectures
    if all_archs:
        ax.legend(fontsize=9, frameon=False, loc="upper left", bbox_to_anchor=(1, 1))

    fig.tight_layout()
    fig.savefig(OUTPUT, dpi=150, bbox_inches="tight")
    gha("notice", f"  Saved to {OUTPUT}")


if __name__ == "__main__":
    main()
