# Ticker Search Tool

This tool helps you find the correct ticker symbols for Tinkoff Invest API instruments.

## Usage

```bash
go run main.go <search_query> [instrument_type]
```

### Examples

```bash
# Search for CNY-related currency instruments
go run main.go "CNY" currency

# Search for SBER shares
go run main.go "SBER" share

# Search for USD instruments without filtering by type
go run main.go "USD"

# Search for Apple shares
go run main.go "Apple" share

# Search for government bonds
go run main.go "ОФЗ" bond
```

### Instrument Types

- `share` or `stock` - Акции (Shares)
- `bond` - Облигации (Bonds)
- `etf` - ETF (Exchange Traded Funds)
- `currency` - Валюты (Currencies)
- `futures` or `future` - Фьючерсы (Futures)
- `option` - Опционы (Options)

## Setup

1. Make sure you have the `TINKOFF_TOKEN` environment variable set:
   ```bash
   export TINKOFF_TOKEN="your_token_here"
   ```

2. Run the tool from the `tinkoff/examples/ticker_search` directory:
   ```bash
   cd tinkoff/examples/ticker_search
   go run main.go "your_search_query"
   ```

## Output

The tool will show you:
- **Name**: Full instrument name
- **Ticker**: The ticker symbol to use in your configuration
- **Class Code**: Exchange class code
- **FIGI**: Financial Instrument Global Identifier
- **Full Ticker**: Combined class code and ticker
- **API Trading Status**: Whether the instrument is available for API trading
- **IIS Availability**: Whether the instrument is available for Individual Investment Accounts
- **Lot Size**: Minimum trading lot size

## Common Use Cases

### Finding Currency Pairs

```bash
# Find CNY/RUB pairs
go run main.go "CNY" currency

# Find USD/RUB pairs
go run main.go "USD" currency
```

### Finding Futures

```bash
# Find CNY futures
go run main.go "CNY" futures

# Find oil futures
go run main.go "нефть" futures
```

### Finding Stocks

```bash
# Find Sberbank shares
go run main.go "SBER" share

# Find Gazprom shares
go run main.go "Газпром" share
```

## Troubleshooting

If you get "instrument not found" errors in your trading application:

1. Use this tool to search for the correct ticker
2. Make sure the instrument has `✅ Available for API trading`
3. Use the exact `Ticker` value shown in the output
4. For some instruments, you might need to use the `FIGI` instead of the ticker

## Note

This tool uses the **demo environment** by default for safety. The same tickers should work in both demo and production environments, but always verify in demo first before using real money.
