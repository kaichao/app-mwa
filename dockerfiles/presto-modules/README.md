
# 说明
目前经过拆分后，原presto-search模块分为rfifind,dedisp,search-fold三个模块。各模块功能如下：

- rfifind: fits -> mask 找出射电干扰。输入消息为待处理fits文件所在目录路径，可以一次处理连续的多个fits文件。
- dedisp: fits + mask -> dat 根据指定的LINENUM环境变量执行一组消色散。目前的方案将所有DM分为9组执行，共6041个DM值。输入消息为待处理fits文件所在目录路径，可以一次处理连续的多个fits文件。程序使用环境变量LINENUM确定执行的消色散参数，产生大量.dat文件。使用zstd对处理结果进行压缩，压缩率约为70%。该模块处理得到的结果为$DIR_DEDISP/${m}/${LINENUM}.tar.zst压缩包。根据现有测试估计，4800s数据产生的.dat文件规模共172G，经过压缩后约为120G。这一规模的数据仍难以存放在内存中，因此仍需要考虑：避免一次生成全部.dat文件，转而分多次生成.dat文件；或者将产生的dat压缩文件包传输至其他节点完成后续计算。
- search-fold: 对输入的一组.dat文件完成以下步骤
    - dat -> fft
    - fft -> cand
    - cand + dat -> 图片（最终结果）

    输入消息为待处理.dat文件压缩包的路径。处理结果为若干干.ps/.pfd/.png文件，和一个包含筛选后结果的candidates.txt文件。

- presto-search-raw: 最早版本的脚本。输入消息为单个fits.zst文件路径，完成整个presto模块的处理流程，并使用原始fits数据执行prepfold。
- presto-search-dat: 输入消息为单个fits.zst文件路径，完成整个presto模块的处理流程，并使用消色散后得到的dat文件执行prepfold。
