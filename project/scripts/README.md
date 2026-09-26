# Script Status

The active scripts in this folder are retained only for data acquisition and data inventory.

Active or reusable:

- `pull_bitget.py`: Bitget public candle puller.
- `process_yahoo.py`: converts previously exported Yahoo Finance data to parquet.
- `pull_fred.py`: FRED daily macro puller.
- `build_events.py`: event-calendar construction.
- `qc_manifest.py`: raw data manifest generation.

Archived as invalid:

- `stage1_segmented_regression.py`
- `build_master_table.py`
- `stage234_analysis.py`

Those scripts are under `../archive_invalid_2026-07-02/scripts/` because they generated invalid IV, mediation, and causal-report outputs.
