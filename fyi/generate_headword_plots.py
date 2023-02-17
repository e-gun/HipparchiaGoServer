import pandas as pd
import matplotlib.pyplot as plt

# latin headwords:
# psql -d hipparchiaDB -o ~/tmp/out.csv -F '|' -A \
#        -c "SELECT entry_name,total_count from dictionary_headword_wordcounts where lt_count > 0 and entry_name ~ '[a-z]' order by total_count desc"

# greek headwords:
# psql -d hipparchiaDB -o ~/tmp/out.csv -F '|' -A \
#        -c "SELECT entry_name,total_count from dictionary_headword_wordcounts where gr_count > 0 and entry_name ~ '[^a-z]' order by total_count desc"

df = pd.read_csv('~/tmp/out.csv', sep='|')

print("median:\t"+str(df['total_count'].median()))
print("mean:\t"+str(df['total_count'].mean()))

df.plot(kind='line', x='entry_name', y='total_count', logy=True)
plt.show()
