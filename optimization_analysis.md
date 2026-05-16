# 2CRT 模擬系統優化與語言轉換分析

目前模擬 $M=120$ 需要耗費高達 28 小時，這在蒙地卡羅模擬（10萬次迭代）中是非常常見的效能瓶頸。MATLAB 雖然在矩陣運算上強大，但若在大量的 `for` 迴圈中頻繁進行記憶體配置與高複雜度函數呼叫，效能會急遽下降。

以下為針對此專案的優化分析，分為「MATLAB 重構建議」、「重構理由」以及「改用 Golang 的優劣分析」。

---

## 1. 如何重構 (MATLAB 內部優化)

### (A) 從「陣列位移」改為「純索引運算」 (Index-based Operations)
*   **現況**：目前程式使用 `circshift` 對長度為 $p \times q$ 的 0/1 陣列進行位移，再用 `find` 找出 `1` 的位置。
    ```matlab
    tmp = circshift(SgSet{g}, random_shift);
    nbo = find(tmp == 1);
    ```
*   **優化**：完全廢棄 0/1 陣列！只記錄「值為 1 的索引 (Index)」。
    當序列需要位移 `S` 時，新索引就是 `(原始索引 + S) mod (p*q)`。
    **這樣就不需要建立巨大的陣列，也不需要在迴圈中執行昂貴的 `circshift` 和 `find`。**

### (B) 捨棄昂貴的陣列相加與 `intersect`
*   **現況**：
    ```matlab
    addSg = addSg + randomSgSet{i}; % 長度為 p*q 的向量相加
    rboth = find(addSg > 0 & addSg <= r);
    sgSuccessCounts{i} = intersect(rboth-1, indexCell{sequences(i)}); % 非常慢
    ```
*   **優化**：
    既然我們已經改用「索引」，我們只需將所有序列當前的「發射索引」收集起來，然後計算哪些索引的「出現次數 $\le r$」。可以使用 MATLAB 的頻率統計技巧（例如 `histcounts` 或 `sort` + 邏輯運算）找出這些合法的索引，然後再過濾出各別序列的成功傳輸。這能避開底層非常耗時的 `intersect`。

### (C) 優化輔助函數 `getOneIndextmp.m` (向量化)
*   **現況**：使用 `for` 迴圈尋找，並在迴圈內動態擴充陣列 `one_Index = [one_Index, ...] `。動態擴充陣列在 MATLAB 裡是大忌，會不斷重新分配記憶體。
*   **優化**：完全不使用迴圈。因為目標是找出每個 $T$ 區間的「第一個」成功封包。可以利用整數除法算出分組 ID：
    ```matlab
    group_idx = floor(S_Index / T);
    % 利用 unique 的 'stable' 屬性，直接抓出每個分組第一次出現的 index
    [~, first_occurrence_idx] = unique(group_idx, 'stable');
    one_Index = S_Index(first_occurrence_idx);
    ```

---

## 2. 重構理由與 Trade-off

*   **避免記憶體重複分配 (Memory Allocation Overhead)**：
    蒙地卡羅模擬會執行 10 萬次。若每次都在迴圈內宣告新的 cell array、建立 $p \times q$ 的 zero array、或是動態擴充陣列長度，作業系統必須不斷為其尋找記憶體區塊，這佔據了執行時間的絕大比例。改為純索引運算與預先配置 (Preallocation) 能根除此問題。
*   **降低函數呼叫成本 (Function Call Overhead)**：
    在 MATLAB 中呼叫自訂函數（如 `getOneIndextmp`）或是複雜底層函數（如 `intersect`、`circshift`）皆有額外的 context switch 負擔。透過向量化與索引數學運算，我們將邏輯推遲給底層高度優化的 C/C++ 引擎來一次性處理。
*   **Trade-off**：
    上述優化會犧牲些微的「程式碼直觀性」。0/1 陣列的相加對人類來說很好理解，但對電腦來說運算量大；改為模數運算與索引統計，需要開發者具備更好的邏輯思維，但換來的是 **10 倍以上的效能提升**。

---

## 3. 程式語言更換分析：改用 Golang 會更快嗎？

**結論：會非常顯著地變快！有可能將 28 小時的模擬時間壓縮到 30 分鐘甚至更短。**

### Golang 的優勢 (Pros)
1. **極致的編譯執行效能**：Golang 是編譯語言 (Compiled Language)，其原生的 `for` 迴圈執行速度遠勝 MATLAB 的直譯執行。這類單純的整數、陣列、迴圈運算，正是 Go 最擅長的領域。
2. **輕量級的併發處理 (Goroutines)**：MATLAB 的 `parpool` (Worker) 是屬於多進程 (Process)，啟動慢且記憶體消耗大。Golang 的 Goroutine 搭配 `sync.WaitGroup` 極度輕量，你能輕鬆將 10 萬次迭代精準切割給 16 個或 32 個 CPU 核心，幾乎沒有額外的記憶體開銷。
3. **嚴謹的型別與記憶體管理**：Go 允許使用 `make` 預先配置 slice 的容量，搭配高效的垃圾回收機制 (GC)，在大規模模擬中不會產生 MATLAB 常見的記憶體碎片化或越跑越慢的問題。

### Golang 的劣勢與挑戰 (Cons & Challenges)
1. **開發轉換成本**：需要將現有的 MATLAB 邏輯完整改寫為 Go 語言。
2. **缺乏科學運算生態系**：MATLAB 有現成的 `circshift`, `intersect`, `unique` 可以直接叫用。在 Go 中，這些邏輯（如：計算陣列交集）都必須自己手寫邏輯實作。不過幸好此專案的數學並不複雜，皆為基本的邏輯判斷與迴圈。
3. **資料輸出差異**：MATLAB 可以一行 `writematrix` 直接輸出 `.xlsx`。在 Go 中，若要輸出 `.xlsx` 需要引入第三方套件（如 `excelize`），或者最簡單的方式是將結果寫成 `.csv` 格式，之後再用 Excel 開啟。

### 總結建議

*   **如果你只是想稍微加快速度，趕快跑完這次實驗**：
    建議**不要換語言**，直接依照上述「第 1 點」將 MATLAB 的陣列替換為「索引運算」並實作「向量化」。這大概只需要改 20~30 行程式碼，卻能大幅縮短執行時間（可能降至幾小時）。
*   **如果這是你未來的核心研究，需要反覆測試各種參數**：
    **強烈建議花 1~2 天時間將其改寫為 Golang**。一旦轉換完成，未來無論參數怎麼調整，都能體驗到飛一般的速度，長痛不如短痛。
