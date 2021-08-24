## fastq_count

counting fastq(https://en.wikipedia.org/wiki/FASTQ_format) reads, bases, N Bases, Q20, Q30, GC
with high performance


#### 1. Installation
```bash
go install github.com/d2jvkpn/fastq_count
```

#### 2. Usage

$ fastq_count  [-phred value]  [-output out.tsv]  <input1.fastq input2.fastq.gz>
  output (tsv) header: Total reads  Total bases  N bases  Q20  Q30  GC
  
Note:

- When input is -, read standard input;

- "pigz -dc *.fastq.gz | fastq_count -" is recommended for gzipped file(s).

    -output string: output summary to a tsv file, default: stdout

    -phred int: set phred value (default 33)

- output example (tsv):

| Reads       | Bases       | N-bases | GC       | Q20       | Q30       |
| ----------- | ----------- | ------- | -------- | --------- | --------- |
| 1.00 M      | 0.15 G      | 0.00%   | 44.20%   | 96.69%    | 91.59%    |
| 1000000     | 150000000   | 5099    | 66294072 | 145037238 | 137378352 |
