#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = [
#   "httpx",
#   "matplotlib",
# ]
# ///

import os
import httpx
import matplotlib.pyplot as plt
import matplotlib.ticker as ticker
import sys
from pathlib import Path

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

    stats = [
        {
            "tag": r["tag_name"],
            "downloads": sum(a["download_count"] for a in r["assets"]),
        }
        for r in releases
    ]
    stats = [s for s in stats if s["downloads"] > 0]
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
    filtered = sorted(filtered, key=lambda s: s["tag"])

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

    colors = []
    for s in filtered:
        if s["tag"] == latest["tag"] and not is_latest_in_top20:
            colors.append("#73726c")  # gray — latest but not top 20
        else:
            colors.append("#378ADD")  # blue — top 20

    bars = ax.bar(tags, counts, color=colors, width=0.7)

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
    ax.annotate(
        f"{counts[top_idx]:,}",
        xy=(top_idx, counts[top_idx]),
        xytext=(0, 4),
        textcoords="offset points",
        ha="center",
        fontsize=8,
    )

    # Legend if latest is shown outside top 20
    if not is_latest_in_top20:
        from matplotlib.patches import Patch

        ax.legend(
            handles=[
                Patch(color="#378ADD", label="top 20 by downloads"),
                Patch(color="#73726c", label="latest release"),
            ],
            fontsize=9,
            frameon=False,
        )

    fig.tight_layout()
    fig.savefig(OUTPUT, dpi=150, bbox_inches="tight")
    gha("notice", f"  Saved to {OUTPUT}")


if __name__ == "__main__":
    main()
