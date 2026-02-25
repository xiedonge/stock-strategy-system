#!/usr/bin/env python3
import argparse
import json
import os
import sqlite3
import sys
from datetime import datetime, timedelta
from typing import Dict, List, Tuple

try:
    import akshare as ak
    import pandas as pd
except Exception as exc:  # pragma: no cover - runtime dependency check
    print("Missing Python dependency: akshare (and pandas).", file=sys.stderr)
    print("Please run ./scripts/sync_akshare.sh to install them.", file=sys.stderr)
    raise SystemExit(1) from exc


def log(message: str) -> None:
    print(message, file=sys.stderr)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Sync A-share data from AkShare into SQLite")
    parser.add_argument("--db", default=os.getenv("DB_PATH", "backend/data/stock.db"))
    parser.add_argument("--mode", default="all", choices=["daily", "minute", "all"])
    parser.add_argument("--symbols", default="", help="Comma-separated stock codes, e.g. 000001,600519")
    parser.add_argument("--limit", type=int, default=50, help="Limit number of stocks when symbols not provided")
    parser.add_argument("--start-date", default="", help="Daily start date YYYYMMDD")
    parser.add_argument("--end-date", default="", help="Daily end date YYYYMMDD")
    parser.add_argument("--min-start", default="", help="Minute start datetime YYYY-MM-DD HH:MM:SS")
    parser.add_argument("--min-end", default="", help="Minute end datetime YYYY-MM-DD HH:MM:SS")
    parser.add_argument("--period", default="30", help="Minute period: 1,5,15,30,60")
    return parser.parse_args()


def ensure_schema(conn: sqlite3.Connection) -> None:
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS stocks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            code TEXT UNIQUE,
            name TEXT,
            exchange TEXT,
            created_at DATETIME,
            updated_at DATETIME
        )
        """
    )
    conn.execute(
        """
        CREATE TABLE IF NOT EXISTS k_lines (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            stock_code TEXT,
            interval TEXT,
            time DATETIME,
            open REAL,
            high REAL,
            low REAL,
            close REAL,
            volume REAL,
            created_at DATETIME
        )
        """
    )
    conn.execute(
        """
        CREATE INDEX IF NOT EXISTS idx_k_lines_code_interval_time
        ON k_lines(stock_code, interval, time)
        """
    )
    conn.commit()


def infer_exchange(code: str) -> str:
    if code.startswith("6"):
        return "SH"
    if code.startswith("0") or code.startswith("3"):
        return "SZ"
    if code.startswith("8") or code.startswith("4") or code.startswith("9"):
        return "BJ"
    return ""


def get_stock_list(limit: int) -> List[Dict[str, str]]:
    df = ak.stock_info_a_code_name()
    df = df.rename(columns={"item": "code", "value": "name"})
    if limit and len(df) > limit:
        df = df.head(limit)
    rows = df.to_dict(orient="records")
    for row in rows:
        row["exchange"] = infer_exchange(row["code"])
    return rows


def upsert_stocks(conn: sqlite3.Connection, stocks: List[Dict[str, str]]) -> None:
    if not stocks:
        return
    now = datetime.utcnow().isoformat(sep=" ", timespec="seconds")
    payload = [
        (row["code"], row.get("name", ""), row.get("exchange", ""), now, now)
        for row in stocks
    ]
    conn.executemany(
        """
        INSERT INTO stocks (code, name, exchange, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?)
        ON CONFLICT(code) DO UPDATE SET
            name=excluded.name,
            exchange=excluded.exchange,
            updated_at=excluded.updated_at
        """,
        payload,
    )
    conn.commit()


def normalize_daily(df: pd.DataFrame) -> pd.DataFrame:
    mapping = {
        "日期": "time",
        "开盘": "open",
        "收盘": "close",
        "最高": "high",
        "最低": "low",
        "成交量": "volume",
    }
    df = df.rename(columns=mapping)
    for key in mapping.values():
        if key not in df.columns:
            raise ValueError(f"Missing column {key} in daily data")
    df = df[list(mapping.values())]
    df["time"] = pd.to_datetime(df["time"]).dt.strftime("%Y-%m-%d")
    return df


def normalize_minute(df: pd.DataFrame) -> pd.DataFrame:
    mapping = {
        "时间": "time",
        "开盘": "open",
        "收盘": "close",
        "最高": "high",
        "最低": "low",
        "成交量": "volume",
    }
    df = df.rename(columns=mapping)
    for key in mapping.values():
        if key not in df.columns:
            raise ValueError(f"Missing column {key} in minute data")
    df = df[list(mapping.values())]
    df["time"] = pd.to_datetime(df["time"]).dt.strftime("%Y-%m-%d %H:%M:%S")
    return df


def delete_existing(conn: sqlite3.Connection, code: str, interval: str, start: str, end: str) -> None:
    conn.execute(
        """
        DELETE FROM k_lines
        WHERE stock_code = ? AND interval = ? AND time BETWEEN ? AND ?
        """,
        (code, interval, start, end),
    )


def insert_klines(conn: sqlite3.Connection, code: str, interval: str, df: pd.DataFrame) -> int:
    if df.empty:
        return 0
    start = df["time"].min()
    end = df["time"].max()
    delete_existing(conn, code, interval, start, end)

    now = datetime.utcnow().isoformat(sep=" ", timespec="seconds")
    payload = [
        (code, interval, row.time, row.open, row.high, row.low, row.close, row.volume, now)
        for row in df.itertuples(index=False)
    ]
    conn.executemany(
        """
        INSERT INTO k_lines (stock_code, interval, time, open, high, low, close, volume, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        """,
        payload,
    )
    return len(payload)


def default_daily_range() -> Tuple[str, str]:
    end = datetime.utcnow().date()
    start = end - timedelta(days=365)
    return start.strftime("%Y%m%d"), end.strftime("%Y%m%d")


def default_minute_range() -> Tuple[str, str]:
    end = datetime.utcnow().date()
    start = end - timedelta(days=20)
    return (
        start.strftime("%Y-%m-%d") + " 09:30:00",
        end.strftime("%Y-%m-%d") + " 15:00:00",
    )


def main() -> None:
    args = parse_args()
    db_path = args.db

    if not os.path.isabs(db_path):
        db_path = os.path.join(os.getcwd(), db_path)

    db_dir = os.path.dirname(db_path)
    if db_dir:
        os.makedirs(db_dir, exist_ok=True)
    conn = sqlite3.connect(db_path)
    ensure_schema(conn)

    symbols = [s.strip() for s in args.symbols.split(",") if s.strip()]
    if not symbols:
        stocks = get_stock_list(args.limit)
        symbols = [row["code"] for row in stocks]
        upsert_stocks(conn, stocks)
    else:
        stocks = [{"code": code, "name": code, "exchange": infer_exchange(code)} for code in symbols]
        upsert_stocks(conn, stocks)

    daily_start, daily_end = default_daily_range()
    if args.start_date:
        daily_start = args.start_date
    if args.end_date:
        daily_end = args.end_date

    min_start, min_end = default_minute_range()
    if args.min_start:
        min_start = args.min_start
    if args.min_end:
        min_end = args.min_end

    summary = {
        "mode": args.mode,
        "stocks": len(symbols),
        "daily_rows": 0,
        "minute_rows": 0,
        "errors": [],
    }

    for idx, code in enumerate(symbols, start=1):
        log(f"[{idx}/{len(symbols)}] syncing {code} ...")
        if args.mode in ("daily", "all"):
            try:
                daily_df = ak.stock_zh_a_hist(
                    symbol=code,
                    period="daily",
                    start_date=daily_start,
                    end_date=daily_end,
                    adjust="",
                )
                daily_df = normalize_daily(daily_df)
                inserted = insert_klines(conn, code, "1d", daily_df)
                summary["daily_rows"] += inserted
            except Exception as exc:  # pragma: no cover
                summary["errors"].append({"symbol": code, "mode": "daily", "error": str(exc)})

        if args.mode in ("minute", "all"):
            try:
                minute_df = ak.stock_zh_a_hist_min_em(
                    symbol=code,
                    start_date=min_start,
                    end_date=min_end,
                    period=str(args.period),
                    adjust="",
                )
                minute_df = normalize_minute(minute_df)
                inserted = insert_klines(conn, code, f"{args.period}m", minute_df)
                summary["minute_rows"] += inserted
            except Exception as exc:  # pragma: no cover
                summary["errors"].append({"symbol": code, "mode": "minute", "error": str(exc)})

        conn.commit()

    conn.close()
    print(json.dumps(summary, ensure_ascii=False))


if __name__ == "__main__":
    main()
