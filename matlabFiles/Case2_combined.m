clc,clear;
delete(gcp('nocreate'));
parpool(6);
% values = [20,21,22,23,24,25];  % 定義 w 的值
% values = [15,16,17,18,19,20];  % 定義 w 的值
% values = [4,17,18,19,20,21];
% T_Values = [1,5,7,35]; ,3,5,11,15,33,55,165

values = [120];
T_Values = [1,6096];

tic
parfor idx = 1:length(values)

    w = values(idx);  % 根據迴圈索引取得對應的值

    for t_idx = 1:length(T_Values)
        T = T_Values(t_idx);  % 取得當前 T 的值

        p = 127;
        r = 5;
        % 1. 讓使用者輸入想選幾個序列
        se = 120;
        q = ceil(2*se/r);
        sequences = [3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,100,101,102,103,104,105,106,107,108,109,110,111,112,113,114,115,116,117,118,119,120,121,122];
        sequenceLength = p*q; 

        % fprintf("T = %d / w = %d", T, w);
        % ✅ **每個 worker 內部都重新定義 `output_dir`**
        output_dir = sprintf('Case2_newcombined_M%dR%d', se, r);  % 確保 worker 知道 output 目錄
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
            gSet = zeros(w, 2);
            Sg = zeros(1, p*q);

            for j = 0:w-1
                gSet(j+1, :) = [mod(g*j, p), mod(j, q)];
            end

            for t = 0:p*q-1
                St = [mod(t, p), mod(t, q)];
                if any(ismember(gSet, St, 'rows'))
                    Sg(t+1) = 1;
                else
                    Sg(t+1) = 0;
                end
            end

            SgSet{g+1} = Sg;
        end

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

        %% 跑loop
        ns = 100000;
        
        ta = zeros(1, se+1);             % 給 AoI
        totalDelay = zeros(1, se+1);     % 給 Delay
        totalFail = zeros(1, se+1);      % 給 Fail
        totalGroupDelay = zeros(1, se+1);% 給 GroupDelay
        totalGroupFail = zeros(1, se+1); % 給 GroupFail
        throughput = zeros(1, se+1);     % 給 Throughput

        for m = 1:ns
            randomSgSet = cell(se, 1);
            indexCell = cell(se, 1);

            for g = 1:p+1
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

            sgSuccessCounts = cell(se,1);

            for i = 1:se
                sgSuccessCounts{i} = intersect(rboth-1, indexCell{sequences(i)});
                throughput(i) = throughput(i) + length(sgSuccessCounts{i});
            end

            %% ======= AoI Calculation =======
            for i = 1:length(sgSuccessCounts)
                if ~isempty(sgSuccessCounts{i})
                    one_Index = getOneIndextmp(sgSuccessCounts{i}, T, p, q);
                    second_group = one_Index + p*q;
                    extended_Index = [one_Index, second_group];

                    extended_AoI = zeros(1,2*p*q);

                    for d = 1:2*p*q
                        if ismember(d,extended_Index)
                            extended_AoI(d) = mod(d,T)+ 1;
                        elseif d == 1
                            extended_AoI(d) = 0;
                        else
                            extended_AoI(d) = extended_AoI(d-1) + 1;
                        end
                    end

                    calculate_AoI = extended_AoI(p*q+1:2*p*q);
                    AoI_average = sum(calculate_AoI)/(p*q);
                    ta(i) = ta(i) + AoI_average;
                end
            end

            %% ======= Delay Calculation =======
            delay = cell(se, 1);
            delayCount = zeros(1,se);
            % failCount = zeros(1,se);
            % groupFailCount = 0;
            % groupDelaySum = 0;

            for g = 1:(sequenceLength/T)
                % 記錄 GroupDelay
                currentMaxDelay = 0;
                currentFail = false;

                for i = 1:length(sgSuccessCounts)
                    % 定義區間範圍: (T*(g-1)) < x <= (T*g)
                    lower_bound = T * (g - 1);
                    upper_bound = T * g;

                    % 尋找符合範圍的第一個值 ('first')
                    idx_loc = find(sgSuccessCounts{i} >= lower_bound & sgSuccessCounts{i} < upper_bound, 1, 'first');

                    if ~isempty(idx_loc)
                        % 如果有找到值
                        first_val = sgSuccessCounts{i}(idx_loc);
                        modValue = mod(first_val, T);
                        if modValue == 0
                            delay{i}(g) = 1;
                        else
                            delay{i}(g) = modValue + 1 ;
                        end

                        % 計算 Sp 的 delayCount
                        delayCount(i) = delayCount(i) + 1;

                        % 記錄 GroupDelay (更新當前組別的最大延遲)
                        currentMaxDelay = max(currentMaxDelay, delay{i}(g));
                    else
                        % 如果沒找到記錄0
                        % delay{i}(g) = 0;
                        % currentFail=true;

                        % 計算 Sp 的 failCount
                        % failCount(i) = failCount(i) + 1;
                    end
                end

                % 把 GroupDelay/fail 記錄起來
                % if currentFail
                %     groupFailCount = groupFailCount + 1;
                % else
                %     groupDelaySum = groupDelaySum + currentMaxDelay;
                % end
            end

            currentDelay = 0;
            currentDelayFail = 0;
            % 累加總 delay & fail
            for i = 1:length(sgSuccessCounts)
                if delayCount(i) ~= 0
                    currentDelay = currentDelay + sum(delay{i})/delayCount(i);
                    totalDelay(i) = totalDelay(i) + sum(delay{i})/delayCount(i);
                end
                % currentDelayFail = currentDelayFail + failCount(i)/(sequenceLength/T);
                % totalFail(i) = totalFail(i) + failCount(i)/(sequenceLength/T);
            end
            totalDelay(se+1) = totalDelay(se+1) + currentDelay/se;
            % totalFail(se+1) = totalFail(se+1) + currentDelayFail/se;

            % 計算 GroupDelay
            % if sequenceLength/T-groupFailCount == 0
            %     groupDelay = 0;
            % else
            %     groupDelay = groupDelaySum / (sequenceLength/T-groupFailCount);
            % end
            % totalGroupDelay(1) = totalGroupDelay(1) + groupDelay;

            % 計算 GroupFail
            % groupFail = groupFailCount / (sequenceLength/T);
            % totalGroupFail(1) = totalGroupFail(1) + groupFail;

        end

        % 打開文件，儲存檔案到對應的資料夾
        filename = fullfile(T_folder, sprintf('%d結果%d.xlsx', T, w));

        % 計算 10萬次的平均
        tmpm = p*q*ns;

        taavg = ta/ns;
        taavg(length(taavg))= sum(ta) / (se*ns);

        avgtd = totalDelay/ns;
        % avgtf = totalFail/ns;
        % avgtgd = totalGroupDelay/ns;
        % avgtgf = totalGroupFail/ns;

        avgthroughput = throughput/tmpm;
        avgthroughput(se+1) = sum(throughput(1:se)) / (se*tmpm);

        % 合併為一個 6 列 (6xN) 的矩陣
        mat = [avgthroughput; avgtd; taavg];
        % mat = [taavg; avgtd; avgtf; avgtgd; avgtgf; avgthroughput];
        writematrix(mat, filename);
    end
end
toc
