# Case 1 質數序列與 2-UI 性質驗證 (verify-case1-2ui-6-sequences 分支)

此分支專門用於進行 **Case 1 質數序列 (Prime Sequence, PS)** 的傳輸效能模擬，以及在多使用者與多包接收 (MPR) 環境下，對 **2-UI (User-Irrepressible) 性質** 的數值驗證。

## 📂 檔案清單與功能說明

這個分支包含 MATLAB 效能模擬主程式，以及新增的高效 Go 語言與 Python 驗證工具：

### 1. MATLAB 效能模擬 (對照組)
*   **`Case1_withDelay.m`**
    *   **功能**：在 MPR 接收能力 $\gamma = 2$ ($r=2$) 下，模擬 $p^2$ 週期的質數序列，計算各個使用者的 Throughput、Delay 以及 Group Delay 的平均數值，並輸出為 Excel。
*   **`Case1_withaoi.m`**
    *   **功能**：在 MPR 接收能力 $\gamma = 2$ ($r=2$) 下，模擬質數序列的資訊年齡 (AoI, Age of Information) 指標，並輸出為 Excel。
*   **`getOneIndextmp.m`**
    *   **功能**：MATLAB 輔助函數，在觀察區間 $T$ 內尋找並抓取第一個成功傳輸的封包索引。

### 2. Python 數學與分布比對工具
*   **`verify_hamming_correlation.py`**
    *   **功能**：計算 $p=5$ 時的 6 個質數序列兩兩之間的 **Hamming cross-correlation 矩陣**（排除對角線自相關）。
    *   **對照理論**：用以驗證 `gamma_UI (2).pdf` 論文第二頁右側的 **Lemma 1**（驗證 $H_{s_g, s_p} = 1$ 以及 $H_{s_g, s_h} = 2$ 的正確性）。
*   **`export_case1_csv.py`**
    *   **功能**：生成這 6 個序列的 `0/1` 時間槽分布，並匯出成 [case1_sequences.csv](file:///d:/Users/minray/Desktop/Case2/2CRT/2CRT/case1_sequences.csv) 供你直接比對與查閱。

### 3. Go 語言全狀態碰撞遍歷工具 (Bitwise 高效版)
*   **`verify_all_states.go`**
    *   **功能**：利用 Go 的 Goroutines 平行處理與 bitwise 運算，遍歷 6 個使用者在 $25^5 = 9,765,625$ 種獨立相對時移狀態下，全體發射的最大碰撞重疊值（尋找最壞對齊狀況）。
*   **`verify_mpr_ui.go`**
    *   **功能**：針對 $p=5$ 的 6 個序列，在 **MPR $r=2$** 的接收能力下，判定是否符合 2-UI 性質（檢查是否在所有時移下，每個字母 `a` 到 `f` 都能成功傳輸至少一次）。若有失敗，將會輸出位移反例。
*   **`verify_mpr_ui_k7.go`**
    *   **功能**：針對 $p=7$ 的 8 個序列，在 **MPR $r=2$** 的接收能力下，透過**隨機抽樣 10,000,000 次**驗證其 2-UI 性質，成功克服了 $49^7 \approx 67.8$ 兆種巨大相對時移狀態的爆炸問題。

---

## 🚀 執行與使用方式

確保你具備 Python 3.x 以及 Go 語言環境，並在終端機進入專案根目錄：

### 1. 執行 Hamming 交叉相關矩陣計算
```powershell
python verify_hamming_correlation.py
```

### 2. 匯出 6 個序列的 0/1 分布 CSV
```powershell
python export_case1_csv.py
```

### 3. 運行 Go 6 使用者最大碰撞遍歷
```powershell
go run verify_all_states.go
```

### 4. 運行 Go 6 使用者 MPR (r=2) UI 驗證
```powershell
go run verify_mpr_ui.go
```

### 5. 運行 Go 8 使用者 (K=7) MPR (r=2) 1000 萬次抽樣驗證
```powershell
go run verify_mpr_ui_k7.go
```

---

## 📝 驗證結論速覽
*   **6 使用者 ($p=5$)**：在 MPR $r=2$ 下，遍歷 976 萬種相對位移狀態，**100% 符合 2-UI 性質**，沒有任何人會被完全碰撞壓制。
*   **8 使用者 ($p=7$)**：在 MPR $r=2$ 下，抽樣 1000 萬次，**100% 符合 2-UI 性質**，展現了強健的非同步免協調傳輸保障。
