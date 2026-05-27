% delete(gcp('nocreate'));
% parpool(6);

clc;

values = [5];  % 定義 w 的值
T_Values = [25];
p = 5;
q = 5;
r = 2;
sequenceLength = p*q;

tic
for idx = 1:length(values)

    w = values(idx);  % 根據迴圈索引取得對應的值

    for t_idx = 1:length(T_Values)
        T = T_Values(t_idx);  % 取得當前 T 的值

        % 1. 讓使用者輸入想選幾個序列
        se = 5;
        sequences = [2,3,4,5,6];

        % fprintf("T = %d / w = %d", T, w);
        % ✅ **每個 worker 內部都重新定義 `output_dir`**
        output_dir = sprintf('Case1_delay_M%dR%d', se, r);  % 確保 worker 知道 output 目錄
        T_folder = fullfile(output_dir, sprintf('T_%d', T));  % T 值的資料夾

        % ✅ **在 worker 內部確認目錄是否存在**
        if exist(output_dir, 'dir') == 0
            mkdir(output_dir);
        end
        if exist(T_folder, 'dir') == 0
            mkdir(T_folder);
        end
        %tic
        %% 做出Sgset
        SgSet = cell(p+1, 1);

        for g = 0:p-1
            gSet = zeros(w, 1);
            Sg = zeros(1, p*p);

            for j = 0:w-1
                gSet(j+1) = mod(j*(1+g*p), p*p);
            end

            for index = 1:length(gSet)
                Sg(gSet(index)+1)=1;
            end

            SgSet{g+1} = Sg;
        end

        Stmp = zeros(1, p*p);
        gSet = zeros(w, 1);
        % Sg = zeros(1, p*p);

        for j = 0:w-1
            gSet(j+1) = mod(j*p, p*p);
        end

        for index = 1:length(gSet)
            Stmp(gSet(index)+1)=1;
        end
        SgSet{p+1} = Stmp;

        Stmp = zeros(1, p*q);
        gSet = zeros(w, 2);
        Sg = zeros(1, p*q);

        for j = 0:w-1
            gSet(j+1, :) = [mod(j, p), mod(0, q)];
        end

        for t = 0:p*q-1
            St = [mod(t, p), mod(t, q)];
            if any(ismember(gSet, St, 'rows'))
                Stmp(t+1) = 1;
            else
                Stmp(t+1) = 0;
            end
        end
        SgSet{p+1} = Stmp;

        % 算 aoi delay fail
        %% 跑loop
        ns = 100000;
        totalDelay = zeros(1, se+1);
        totalFail = zeros(1, se+1);
        totalGroupDelay = zeros(1, se+1);
        totalGroupFail = zeros(1, se+1);

        for m = 1:ns
            randomSgSet = cell(se, 1);
            indexCell = cell(se, 1);

            for g = 1:se+1
                if ismember(g, sequences)
                    random_shift = randi([0, p*q]);
                    tmp = circshift(SgSet{g}, random_shift);
                    randomSgSet{g} = tmp;
                    nbo = find(randomSgSet{g} == 1);
                    indexCell{g} = nbo-1;
                else
                    randomSgSet{g} = zeros(1, p*q);
                end
            end

            addSg = zeros(1, length(SgSet{1}));

            for i = 1:length(randomSgSet)
                addSg = addSg + randomSgSet{i};
            end

            rboth = find(addSg > 0 & addSg <= r);

            sgSuccessCounts = cell(p,1);
            groupDelayTmp = 0;


            for i = 1:se
                sgSuccessCounts{i} = intersect(rboth-1, indexCell{sequences(i)});
            end

            % 算 delay
            delay = cell(se, 1);
            delayCount = zeros(1,se);
            failCount = zeros(1,se);
            groupFailCount = 0;
            groupDelaySum = 0;

            for g = 1:(sequenceLength/T)
                % 記錄 GroupDelay
                currentMaxDelay = 0;
                currentFail = false;

                for i = 1:length(sgSuccessCounts)
                    % 定義區間範圍: (T*(g-1)) < x <= (T*g)
                    lower_bound = T * (g - 1);
                    upper_bound = T * g;

                    % 3. 尋找符合範圍的第一個值 ('first')
                    % 注意：這裡條件要用 & 連接，且變數是 sgSuccessCounts(...)
                    idx_loc = find(sgSuccessCounts{i} >= lower_bound & sgSuccessCounts{i} < upper_bound, 1, 'first');

                    if ~isempty(idx_loc)
                        % 如果有找到值
                        first_val = sgSuccessCounts{i}(idx_loc);
                        modValue = mod(first_val, T);
                        if modValue == 0
                            delay{i}(g) = 1;
                            % delay{i}(g) = 0;
                        else
                            delay{i}(g) = modValue + 1 ;
                            % delay{i}(g) = modValue ;
                        end

                        % 計算 Sp 的 delayCount
                        delayCount(i) = delayCount(i) + 1;

                        % 記錄 GroupDelay (更新當前組別的最大延遲)
                        currentMaxDelay = max(currentMaxDelay, delay{i}(g));
                    else
                        % 如果沒找到記錄0
                        delay{i}(g) = 0;
                        currentFail=true;

                        % 計算 Sp 的 delayCount
                        failCount(i) = failCount(i) + 1;
                    end
                end

                % 把 GroupDelay/fail 記錄起來
                % 假設是 0，代表 group fail+1
                % 有值的話，要把 GroupDelay 加總
                if currentFail
                    groupFailCount = groupFailCount + 1;
                else
                    groupDelaySum = groupDelaySum + currentMaxDelay;
                end

            end

            currentDelay = 0;
            currentDelayFail = 0;
            % 計算 delay
            for i = 1:length(sgSuccessCounts)
                if delayCount(i) ~= 0
                    currentDelay = currentDelay + sum(delay{i})/delayCount(i);
                    totalDelay(i) = totalDelay(i) + sum(delay{i})/delayCount(i);
                end
                % currentDelay = currentDelay + sum(delay{i})/delayCount(i);
                currentDelayFail = currentDelayFail + failCount(i)/(sequenceLength/T);
                % totalDelay(i) = totalDelay(i) + sum(delay{i})/delayCount(i);
                totalFail(i) = totalFail(i) + failCount(i)/(sequenceLength/T);
            end
            totalDelay(p+1) = totalDelay(p+1) + currentDelay/p;
            totalFail(p+1) = totalFail(p+1) + currentDelayFail/p;
            % td = sum(delay{p}) / delayCount;  % 平均延遲
            % fprintf('S%d的平均td: %f\n', p, td);
            % fprintf('%f\n', td);
            % totalDelay = totalDelay + td;

            % 計算 fail
            % fail = failCount/(sequenceLength/T);
            % % fprintf('%f\n', fail);
            % totalFail = totalFail + fail;

            % 計算 GroupDelay
            if sequenceLength/T-groupFailCount ==0
                groupDelay = 0;
            else
                groupDelay = groupDelaySum/ (sequenceLength/T-groupFailCount);
                % fprintf('%f\n', groupDelay);
            end
            totalGroupDelay(1) = totalGroupDelay(1) + groupDelay;

            % 計算 GroupFail
            groupFail = groupFailCount / (sequenceLength/T);
            % fprintf('%f\n', groupFail);
            totalGroupFail(1) = totalGroupFail(1) + groupFail;


            % tmpm = sequenceLength*ns;

            % 打開文件
            % 儲存檔案到對應的資料夾
            % filename = fullfile(T_folder, sprintf('%d結果%d.xlsx', T, w));
            % taavg = ta/ns;
            % taavg(length(taavg))= sum(ta) / (se*ns);
            % mat = taavg;
            % writematrix(taavg, filename);
        end

        filename = fullfile(T_folder, sprintf('%d結果%d.xlsx', T, w));
        % 加總10萬次後平均
        avgtd = totalDelay/ns;
        avgtf = totalFail/ns;
        avgtgd =totalGroupDelay/ns;
        avgtgf = totalGroupFail/ns;

        mat = [avgtd; avgtf; avgtgd; avgtgf];
        writematrix(mat, filename)
    end
    toc

end