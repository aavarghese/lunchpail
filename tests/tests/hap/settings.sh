# hap test
api=ray
app="$TOP"/watsonx_ai/preprocessing/hap
expected=('estimated_memory_footprint')

if which lspci && lspci | grep -iq nvidia; then
    expected+=('running in GPU mode' 'torch.cuda.is_available(): True')
else
    expected+=('running in CPU-only mode')
fi
