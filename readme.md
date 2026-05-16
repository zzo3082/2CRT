# 2CRT 網路排程效能模擬 (Network Scheduling Performance Simulation based on CRT)

此專案為一個使用 MATLAB 撰寫的網路通訊模擬系統，主要基於「中國剩餘定理 (Chinese Remainder Theorem, CRT)」來設計序列，並評估多節點網路環境下的傳輸效能。

## 專案目標

本專案模擬非同步網路環境中，多個節點（使用者）使用 CRT 序列進行資料傳輸的情境。系統考量了「多封包接收能力 (Multi-Packet Reception, MPR)」，即當同一時間槽內發生碰撞的封包數量小於或等於容忍值 $r$ 時，封包仍能成功接收。專案透過蒙地卡羅模擬 (Monte Carlo Simulation)，評估在此機制下的三項關鍵網路效能指標：
1. **Throughput (吞吐量)**
2. **Delay (延遲時間)**
3. **Age of Information, AoI (資訊年齡)**

## 核心檔案說明

- **`Case2_combined.m`**
  - **功能**：專案的主程式。負責產生 CRT 序列、建立平行運算池 (`parpool`) 進行 10 萬次迭代的模擬實驗，並計算各項效能指標。
  - **運作流程**：
    1. 產生基於參數 $p$ 與 $q$ 的 CRT 序列集合。
    2. 透過隨機循環位移 (`circshift`) 來模擬節點間的非同步傳輸 (Asynchronous transmission)。
    3. 統計每個時間槽內的傳輸數量，並根據 MPR 能力 (參數 `r`) 判斷封包是否成功接收。
    4. 依據成功接收的紀錄，計算 Throughput、Delay 以及 AoI。
    5. 將最終平均結果輸出並儲存至根據參數動態建立的資料夾中（例如：`Case2_newcombined_M120R5/T_1/` 底下的 Excel 檔案）。

- **`getOneIndextmp.m`**
  - **功能**：輔助函數。
  - **運作流程**：輸入成功接收的封包索引集合 (`S_Index`) 以及觀察區間長度 ($T$)，尋找並回傳每個 $T$ 區間內「第一個」成功傳輸的封包索引。這個函數主要用於輔助計算系統的 Delay 與 AoI。

## 系統參數設定

程式內包含多個可調參數，可用於不同場景的模擬：
- $p$: 素數或序列長度參數 (預設 `127`)。
- $r$: 多封包接收 (MPR) 能力上限 (預設 `5`)。
- $se$: 模擬的使用者數量 (預設 `120`)。
- $q$: 由 $se$ 與 $r$ 動態計算而得，決定完整的序列長度 ($p \times q$)。
- $T$: 觀察區間或容忍延遲參數 (`T_Values` 預設為 `[1, 6096]`)。
- $w$: 用於控制序列生成及內部結構的參數。
- `ns`: 模擬的總次數 (預設 `100,000` 次)。

## 執行方式 (MATLAB 版)

1. 確保電腦已安裝 MATLAB 並具備 Parallel Computing Toolbox（平行運算工具箱）。
2. 在 MATLAB 中開啟此專案資料夾。
3. 執行 `Case2_combined.m`。
4. 程式將自動開啟 `parpool` (預設 6 個 worker) 加速計算，並依據參數將實驗結果匯出至自動建立的 `Case2_newcombined_...` 資料夾內的 `.xlsx` 檔案中。

## Golang 高效能重構版

為了大幅提升蒙地卡羅模擬的執行效率，本專案已重構為 Golang 語言版本。將原本高耗時的陣列平移運算改為「純索引 (Index-based)」運算，並利用 Goroutines 實作了極輕量且高效的平行運算，將原本需要 28 小時以上的實驗大幅縮短至數分鐘內完成。

### Golang 核心檔案說明

- **`main.go`**：主程式進入點。負責定義系統參數 ($p, r, se, T$ 等)，並透過 `sync.WaitGroup` 啟動多個 Goroutines（預設運用所有 CPU 核心）進行平行模擬，最後聚合結果。所有執行日誌會同時輸出至終端機與 `simulation.log` 中。
- **`sequence.go`**：序列處理模組。實作 CRT 序列的生成與隨機位移 (`circshift`)。此模組徹底捨棄了 `0/1` 陣列，只紀錄值為 1 的「索引值」，以極大化運算效率。
- **`simulation.go`**：模擬碰撞與 MPR 核心。收集所有序列的發射索引並統計每個時間槽內的碰撞次數，藉此篩選出符合容忍值 $r$ 的成功傳輸紀錄，取代了原本 MATLAB 中昂貴的 `intersect` 計算。
- **`metrics.go`**：網路指標計算。負責計算每個使用者的吞吐量 (Throughput)、延遲 (Delay) 以及資訊年齡 (AoI)。其中 Delay 的尋找邏輯使用了整數除法分群 ($group = index / T$)，將複雜度優化至 O(1)。
- **`export.go`**：資料輸出模組。將聚合後的平均實驗數據寫入到對應的資料夾與 `.csv` 檔案中。

### 執行方式 (Golang 版)

1. 確保電腦已安裝 [Go 語言環境](https://golang.org/dl/)。
2. 開啟終端機 (Terminal) 並進入此專案資料夾。
3. 如果需要修改參數，請直接編輯 `main.go` 內的系統參數。
4. 編譯並執行程式：
   ```powershell
   go build -o crt_sim.exe
   .\crt_sim.exe
   ```
5. 執行完成後，實驗結果會匯出至 `Case2_newcombined_..._GO/` 資料夾內的 `.csv` 檔案，且執行過程會被記錄在 `simulation.log` 中。
