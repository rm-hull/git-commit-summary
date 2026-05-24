#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = [
#   "httpx",
#   "matplotlib",
# ]
# ///

import httpx
import matplotlib.pyplot as plt
import matplotlib.ticker as ticker
import sys
from pathlib import Path

REPO = "rm-hull/git-commit-summary"
OUTPUT = Path(__file__).parents[2] / "download_stats.png"


def fetch_releases(repo: str) -> list[dict]:
    releases = []
    page = 1
    with httpx.Client() as client:
        while True:
            r = client.get(
                f"https://api.github.com/repos/{repo}/releases",
                params={"per_page": 100, "page": page},
                headers={"Accept": "application/vnd.github+json"},
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
    print(f"Fetching releases for {repo}...")

    releases = fetch_releases(repo)

    stats = [
        {
            "tag": r["tag_name"],
            "downloads": sum(a["download_count"] for a in r["assets"]),
        }
        for r in releases
    ]
    stats = [s for s in stats if s["downloads"] > 0]
    stats.reverse()

    if not stats:
        fig, ax = plt.subplots(figsize=(10, 5))
        ax.text(0.5, 0.5, "No download data yet", ha="center", va="center",
                transform=ax.transAxes, fontsize=13, color="#888")
        ax.set_title(f"Downloads per release — {repo}", fontsize=13, pad=12)
        ax.spines[["top", "right", "left", "bottom"]].set_visible(False)
        ax.set_xticks([])
        ax.set_yticks([])
        fig.tight_layout()
        fig.savefig(OUTPUT, dpi=150, bbox_inches="tight")
        print(f"  No downloads yet, saved placeholder to {OUTPUT}")
        return

    tags = [s["tag"] for s in stats]
    counts = [s["downloads"] for s in stats]
    total = sum(counts)
    top = max(stats, key=lambda s: s["downloads"])

    print(f"  Releases with downloads: {len(stats)}")
    print(f"  Total downloads:         {total:,}")
    print(f"  Top release:             {top['tag']} ({top['downloads']:,} downloads)")

    fig, ax = plt.subplots(figsize=(max(10, len(tags) * 0.4), 5))
    bars = ax.bar(tags, counts, color="#378ADD", width=0.7)

    ax.set_title(f"Downloads per release — {repo}", fontsize=13, pad=12)
    ax.set_xlabel("Release", fontsize=10)
    ax.set_ylabel("Downloads", fontsize=10)
    ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x, _: f"{int(x):,}"))
    ax.tick_params(axis="x", rotation=45, labelsize=8)
    ax.tick_params(axis="y", labelsize=9)
    ax.spines[["top", "right"]].set_visible(False)
    ax.grid(axis="y", color="#e0e0e0", linewidth=0.5)
    ax.set_axisbelow(True)

    # Annotate the top bar
    top_idx = counts.index(max(counts))
    ax.bar_label(
        plt.matplotlib.container.BarContainer([bars[top_idx]]),
        labels=[f"{counts[top_idx]:,}"],
        padding=3,
        fontsize=8,
    )

    fig.tight_layout()
    fig.savefig(OUTPUT, dpi=150, bbox_inches="tight")
    print(f"  Saved to {OUTPUT}")


if __name__ == "__main__":
    main()
