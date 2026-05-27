delete(gcp('nocreate'));
parpool(6);
% values = [20,21,22,23,24,25];  % 定義 w 的值
% values = [15,16,17,18,19,20];  % 定義 w 的值
% values = [4,17,18,19,20,21];
% T_Values = [1,5,7,35]; ,3,5,11,15,33,55,165

values = [13,14,15,16,17,18];
T_Values = [169];

tic
parfor idx = 1:length(values)

    w = values(idx);  % 根據迴圈索引取得對應的值

    for t_idx = 1:length(T_Values)
        T = T_Values(t_idx);  % 取得當前 T 的值

        p = 13;
        q = 13;
        r = 2;

        % 1. 讓使用者輸入想選幾個序列
        se = 13;
        sequences = [2,3,4,5,6,7,8,9,10,11,12,13,14];

        % fprintf("T = %d / w = %d", T, w);
        % ✅ **每個 worker 內部都重新定義 `output_dir`**
        output_dir = sprintf('Case1_ta_output_M%dR%d', se, r);  % 確保 worker 知道 output 目錄
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

        %% 跑loop
        ns = 100000;
        % tt = zeros(1, se+1);
        % td = zeros(1, se+1);
        ta = zeros(1, se+1);
        % sgfail = zeros(1, se+1);
        % groupDelaySum = 0;

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
            groupDelayTmp = 0;


            for i = 1:se
                sgSuccessCounts{i} = intersect(rboth-1, indexCell{sequences(i)});

                % if isempty(sgSuccessCounts{i})
                %     sgfail(i) = sgfail(i) + 1;
                % end
            end

            for i = 1:length(sgSuccessCounts)
                if ~isempty(sgSuccessCounts{i})
                    first_value = sgSuccessCounts{i}(1);

                    if first_value>groupDelayTmp
                        groupDelayTmp = first_value;
                    end

                    % tt(i) = tt(i) + length(sgSuccessCounts{i});
                    % td(i) = td(i) + first_value;

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
            % groupDelaySum = groupDelaySum + groupDelayTmp;

        end

        tmpm = p*q*ns;

        % 打開文件
        % 儲存檔案到對應的資料夾
        filename = fullfile(T_folder, sprintf('%d結果%d.xlsx', T, w));

        % ttavg = tt/tmpm;
        % tdavg = td/ns;
        taavg = ta/ns;
        % sgfailavg = sgfail/ns;

        % ttavg(length(ttavg))= sum(tt) / (se*tmpm);
        % tdavg(length(tdavg))= sum(td) / (se*ns);
        taavg(length(taavg))= sum(ta) / (se*ns);
        % sgfailavg(length(sgfailavg))= sum(sgfail) / (se*ns);

        % groupDelay = zeros(1, se+1);
        % groupDelay(1)=groupDelaySum/ns;

        % mat = [ttavg; tdavg; sgfailavg; groupDelay];
        mat = taavg;
        writematrix(mat, filename);
    end
end
toc

