# run-benchmarks.ps1

param(
    [int]$iterations = 5,
    [string]$output = ""
)

Write-Host "Running benchmarks with $iterations iterations..." -ForegroundColor Green

$timestamp = Get-Date -Format "yyyy-MM-dd_HH-mm-ss"
$outputfile = if ($output) { $output } else { "benchmarks/bench_$timestamp.txt" }

# Ensure the benchmarks directory exists
New-Item -ItemType Directory -Path "benchmarks" -Force | Out-Null

# Run benchmarks with proper escaping
$result = go test "-bench=." -benchmem "-count=$iterations" ./cmd

# Save to output file
$result | Out-File -FilePath $outputfile

Write-Host "`nBenchmark results saved to: $outputfile" -ForegroundColor Cyan
