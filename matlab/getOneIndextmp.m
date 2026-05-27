% CRT Function definition
function one_Index = getOneIndextmp(S_Index, T, p, q)
    one_Index = [];
    numGroups = floor((p*q) / T) + 1;

    for group = 0:(numGroups - 1)
        startIndex = group * T;
        endIndex = (group + 1) * T - 1;
        indicesInGroup = S_Index(S_Index >= startIndex & S_Index <= endIndex);

        if ~isempty(indicesInGroup)
            one_Index = [one_Index, indicesInGroup(1)];
        end
    end
end