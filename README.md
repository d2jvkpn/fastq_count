# FastqCount

Fastq reads, bases, N Bases, Q20, Q30, GC summary with high performance

**Usage**:

summary single fastq(.gz) file:
```./FastqCount input.fastq```

summary multipy files:
```pigz -dc R1.fastq.gz R2.fastq.gz | ./FastqCount -```

**Note**: pipeline stdin make FastqCount faster with gzipped file(s).

**Output** example (tsv):

| Reads | Bases | N-bases | Q20 | Q30 | GC |
| ----------- | ----------- | ------- | --- | --- | -- |
| 1.00 M | 0.15 G | 0.00% | 96.69% | 91.59% | 44.20% |
| 1000000 | 150000000 | 5099 | 145037238 | 137378352 | 66294072 |
