package templates

// GetHTMLTemplate returns the main HTML template
func GetHTMLTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>HFT Backtester</title>
    <script src="https://cdn.jsdelivr.net/npm/uplot/dist/uPlot.iife.min.js"></script>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/uplot/dist/uPlot.min.css">
    <link rel="stylesheet" href="templates/styles.css">
</head>
<body>
    <div class="container">
        <section class="config-section">
            <h2>Backtest Configuration</h2>
            <div class="form-group">
                <label for="hourSelect">Select Hour:</label>
                <select id="hourSelect">
                    <option value="">Loading hours...</option>
                </select>
            </div>
            
            <div class="form-group">
                <label for="strategySelect">Strategy:</label>
                <select id="strategySelect" onchange="updateStrategyParams()">
                    <option value="bollinger">Bollinger Bands</option>
                </select>
            </div>
            
            <div class="form-group">
                <label for="initialCash">Initial Cash:</label>
                <input type="number" id="initialCash" value="100" step="100">
            </div>
            
            <div class="form-group">
                <label for="positionSize">Position Size ($):</label>
                <input type="number" id="positionSize" value="50" step="10">
            </div>
            
            <div class="form-group">
                <label for="commission">Commission (%):</label>
                <input type="number" id="commission" value="0.05" step="0.01" min="0">
            </div>
            
            <div class="form-group">
                <label>Strategy Params:</label>
                <div class="strategy-params" id="strategyParams">
                    <label for="bbPeriod">Period:</label>
                    <input type="number" id="bbPeriod" value="100" min="2">
                    <label for="bbStdDev">Standard Deviations:</label>
                    <input type="number" id="bbStdDev" value="1" step="0.1" min="0.1">
                </div>
            </div>
            
            <button onclick="runBacktest()">Run Backtest</button>
        </section>
        
        <section class="results-section">
            <h2>Backtest Results</h2>
            <div id="status">Ready</div>
            
            <div class="metrics-grid" id="metrics">
                <!-- Metrics will be populated here -->
            </div>
            
            <div class="chart-container">
                <h3>Equity Curve</h3>
                <div id="equityChart"></div>
            </div>
            
            <div id="tradesTable"></div>
        </section>
    </div>
    
    <script src="templates/script.js"></script>
</body>
</html>`
}
