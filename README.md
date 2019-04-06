# FastqCount

Fastq reads, bases, N Bases, Q20, Q30, GC summary with high performance

**Usage**:

  $ FastqCount  [-phred value]  [-o tsv]  <input1.fastq input2.fastq.gz>
  
  output (tsv) header: Total reads  Total bases  N bases  Q20  Q30  GC
  
  note:
  
  1. When input is -, read standard input;
    
  2. "pigz -dc *.fastq.gz | FastqCount -" is recommended for gzipped file(s).

  -o string
  
  output summary to a tsv file, default: stdout
  
  -phred int
  
  set phred value (default 33)

**Output** example (tsv):

| Reads | Bases | N-bases | Q20 | Q30 | GC |
| ----------- | ----------- | ------- | --- | --- | -- |
| 1.00 M | 0.15 G | 0.00% | 96.69% | 91.59% | 44.20% |
| 1000000 | 150000000 | 5099 | 145037238 | 137378352 | 66294072 |
