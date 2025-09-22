package templates

// GetHTMLTemplate returns the main HTML template
func GetHTMLTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>HFT Backtester</title>
    <script src="https://cdn.jsdelivr.net/npm/uplot/dist/uPlot.iife.min.js"></script>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/uplot/dist/uPlot.min.css">
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 20px; 
            background-color: #f5f5f5;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        
        header {
            text-align: center;
            margin-bottom: 30px;
            padding: 20px;
            background-color: #fff;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        
        .config-section {
            background-color: #fff;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        
        .config-section h2 {
            margin-top: 0;
            color: #333;
        }
        
        .form-group {
            margin-bottom: 15px;
        }
        
        label {
            display: inline-block;
            width: 150px;
            font-weight: bold;
            color: #555;
        }
        
        select, input {
            padding: 8px 12px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
        }
        
        button {
            padding: 10px 20px;
            background: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            transition: background 0.3s;
        }
        
        button:hover {
            background: #0056b3;
        }
        
        .results-section {
            background-color: #fff;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        
        .results-section h2 {
            margin-top: 0;
            color: #333;
        }
        
        #status {
            padding: 10px;
            margin: 10px 0;
            border-radius: 4px;
            background-color: #e9ecef;
            color: #495057;
        }
        
        .chart-container {
            margin: 20px 0;
            padding: 15px;
            background-color: #f8f9fa;
            border-radius: 4px;
        }
        
        #priceChart, #equityChart {
            width: 100%;
            height: 500px;
        }
        
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
        }
        
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        
        th {
            background-color: #f2f2f2;
            font-weight: bold;
        }
        
        tr:hover {
            background-color: #f5f5f5;
        }
        
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin: 20px 0;
        }
        
        .metric-card {
            background-color: #e3f2fd;
            padding: 15px;
            border-radius: 4px;
            text-align: center;
        }
        
        .metric-value {
            font-size: 24px;
            font-weight: bold;
            color: #1976d2;
        }
        
        .metric-label {
            font-size: 14px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>HFT Backtester v1.0</h1>
        </header>
        
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
                <input type="number" id="positionSize" value="1000" step="100">
            </div>
            
            <div class="form-group">
                <label for="commission">Commission (%):</label>
                <input type="number" id="commission" value="0.05" step="0.01" min="0">
            </div>
            
            <div class="form-group">
                <label>Strategy Params:</label>
                <div class="strategy-params" id="strategyParams">
                    <label for="bbPeriod">Period:</label>
                    <input type="number" id="bbPeriod" value="20" min="2">
                    <label for="bbStdDev">Standard Deviations:</label>
                    <input type="number" id="bbStdDev" value="2" step="0.1" min="0.1">
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
                <h3>Price Chart with Trade Markers</h3>
                <div id="priceChart"></div>
            </div>
            
            <div class="chart-container">
                <h3>Equity Curve</h3>
                <div id="equityChart"></div>
            </div>
            
            <div id="tradesTable"></div>
        </section>
    </div>
    
    <script>
        let pricePlot = null;
        let equityPlot = null;
        let uplot = null; // Объявляем переменную uplot
        
        // Load available hours
        fetch('/api/hours')
            .then(response => response.json())
            .then(hours => {
                const select = document.getElementById('hourSelect');
                select.innerHTML = '<option value="">Select an hour</option>';
                
                hours.sort((a, b) => a.hour.localeCompare(b.hour));
                hours.forEach(hourInfo => {
                    const option = document.createElement('option');
                    option.value = hourInfo.hour;
                    option.textContent = hourInfo.hour + ':00 (' + hourInfo.count + ' trades)';
                    select.appendChild(option);
                });
                
                document.getElementById('status').textContent = 'Found ' + hours.length + ' hours of data';
            })
            .catch(err => {
                document.getElementById('status').textContent = 'Error loading hours: ' + err.message;
            });
        
        function updateStrategyParams() {
            const strategy = document.getElementById('strategySelect').value;
            const paramsDiv = document.getElementById('strategyParams');
            
            if (strategy === 'bollinger') {
                paramsDiv.innerHTML = '<label for="bbPeriod">Period:</label><input type="number" id="bbPeriod" value="20" min="2"><label for="bbStdDev">Standard Deviations:</label><input type="number" id="bbStdDev" value="2" step="0.1" min="0.1">';
            }
        }
        
        function runBacktest() {
            document.getElementById('status').textContent = 'Running backtest...';
            
            const strategy = document.getElementById('strategySelect').value;
            const initialCash = parseFloat(document.getElementById('initialCash').value);
            const positionSize = parseFloat(document.getElementById('positionSize').value);
            const commission = parseFloat(document.getElementById('commission').value);
            const strategyParams = {};
            
            if (strategy === 'bollinger') {
                strategyParams.period = parseInt(document.getElementById('bbPeriod').value);
                strategyParams.stdDev = parseFloat(document.getElementById('bbStdDev').value);
            }
            
            const requestData = {
                strategy: strategy,
                initial_cash: initialCash,
                position_size: positionSize,
                commission: commission,
                hour: document.getElementById('hourSelect').value,
                strategy_params: strategyParams
            };
            
            fetch('/api/backtest', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(requestData)
            })
            .then(response => response.json())
            .then(data => {
                document.getElementById('status').textContent = 'Backtest completed';
                displayResults(data);
            })
            .catch(err => {
                document.getElementById('status').textContent = 'Error running backtest: ' + err.message;
            });
        }
        
        function displayResults(data) {
            // Display metrics
            const metricsDiv = document.getElementById('metrics');
            const initialEquity = data.initial_cash || 100;
            const finalEquity = data.final_equity;
            const profit = finalEquity - initialEquity;
            const profitPct = (profit / initialEquity) * 100;
            
            metricsDiv.innerHTML = \`
                <div class="metric-card">
                    <div class="metric-value">$${finalEquity.toFixed(2)}</div>
                    <div class="metric-label">Final Equity</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">${data.trades.length}</div>
                    <div class="metric-label">Number of Trades</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">$${profit.toFixed(2)}</div>
                    <div class="metric-label">Profit</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">${profitPct.toFixed(2)}%</div>
                    <div class="metric-label">Profit Percentage</div>
                </div>
            \`;
            
            // Display equity curve
            if (data.equity_curve && data.equity_curve.length > 0) {
                if (equityPlot) {
                    equityPlot.destroy();
                }
                
                const timestamps = data.equity_curve.map(point => point.time / 1000);
                const equity = data.equity_curve.map(point => point.equity);
                
                const opts = {
                    title: "Equity Curve (Initial: $" + (data.initial_cash || 100) + ")",
                    id: "equity-chart",
                    class: "my-chart",
                    width: document.getElementById('equityChart').offsetWidth,
                    height: 500,
                    series: [
                        {},
                        {
                            label: "Equity",
                            stroke: "green",
                            width: 2,
                            fill: "rgba(0, 255, 0, 0.1)",
                            points: { show: false }
                        }
                    ],
                    axes: [
                        {
                            label: "Time",
                            labelSize: 80,
                            stroke: "black",
                            grid: { show: true, stroke: "#eee" },
                            ticks: { show: true, stroke: "#ddd" },
                            scale: "x",
                            values: [
                                [3600000, "{HH}:{mm}", null, null, null, 1],
                                [600, "{HH}:{mm}", null, null, null, 1],
                                [100, "{HH}:{mm}:{ss}", null, null, 1]
                            ]
                        },
                        {
                            label: "Equity ($)",
                            labelSize: 60,
                            stroke: "black",
                            grid: { show: true, stroke: "#eee" },
                            ticks: { show: true, stroke: "#ddd" },
                            scale: "y"
                        }
                    ],
                    cursor: {
                        show: true,
                        drag: { show: true, x: true, y: false, setScale: true }
                    },
                    legend: { show: true }
                };
                
                equityPlot = new uPlot(opts, [timestamps, equity], document.getElementById('equityChart'));
            }
            
            // Display price chart with trade markers
            if (data.price_data && data.price_data.length > 0) {
                if (pricePlot) {
                    pricePlot.destroy();
                }
                
                // Prepare data for uPlot
                const timestamps = data.price_data.map(point => point.time / 1000); // Convert ms to seconds
                const prices = data.price_data.map(point => point.price);
                
                // Configure uPlot
                const opts = {
                    title: "Price Chart with Trade Markers",
                    id: "price-chart",
                    class: "my-chart",
                    width: document.getElementById('priceChart').offsetWidth,
                    height: 500,
                    series: [
                        {},
                        {
                            label: "Price",
                            stroke: "blue",
                            width: 1,
                            fill: "rgba(0, 255, 0.1)",
                            points: { show: false }
                        }
                    ],
                    axes: [
                        {
                            label: "Time",
                            labelSize: 80,
                            stroke: "black",
                            grid: { show: true, stroke: "#eee" },
                            ticks: { show: true, stroke: "#ddd" },
                            scale: "x",
                            values: [
                                [360000, "{HH}:{mm}", null, null, 1],
                                [600, "{HH}:{mm}", null, null, 1],
                                [1000, "{HH:mm:ss}", null, null, 1]
                            ]
                        },
                        {
                            label: "Price",
                            labelSize: 60,
                            stroke: "black",
                            grid: { show: true, stroke: "#eee" },
                            ticks: { show: true, stroke: "#ddd" },
                            scale: "y"
                        }
                    ],
                    cursor: {
                        show: true,
                        drag: { show: true, x: true, y: false, setScale: true },
                        sync: { key: "hft-backtester" }
                    },
                    legend: { show: true }
                };
                
                // Create uPlot chart
                pricePlot = new uPlot(opts, [timestamps, prices], document.getElementById('priceChart'));
                
                // Add trade markers if we have trades
                if (data.trades && data.trades.length > 0) {
                    // Prepare trade marker data
                    const buyTrades = [];
                    const sellTrades = [];
                    
                    data.trades.forEach(trade => {
                        const tradeTime = trade.time / 1000; // Convert ms to seconds
                        if (trade.is_buy) {
                            buyTrades.push([tradeTime, trade.price]);
                        } else {
                            sellTrades.push([tradeTime, trade.price]);
                        }
                    });
                    
                    // Add buy markers (green triangles)
                    if (buyTrades.length > 0) {
                        pricePlot.addSeries({
                            label: "Buy Trades",
                            stroke: "green",
                            fill: "green",
                            points: {
                                show: true,
                                size: 5,
                                symbol: "triangleUp"
                            }
                        }, 2); // Add as third series (index 2)
                    }
                    
                    // Add sell markers (red triangles)
                    if (sellTrades.length > 0) {
                        pricePlot.addSeries({
                            label: "Sell Trades",
                            stroke: "red",
                            fill: "red",
                            points: {
                                show: true,
                                size: 5,
                                symbol: "triangleDown"
                            }
                        }, 3); // Add as fourth series (index 3)
                    }
                    
                    // Update data with trade markers
                    if (buyTrades.length > 0 || sellTrades.length > 0) {
                        const allData = [timestamps, prices];
                        if (buyTrades.length > 0) allData.push(...buyTrades.map(t => [t[0], t[1]]));
                        if (sellTrades.length > 0) allData.push(...sellTrades.map(t => [t[0], t[1]]));
                        pricePlot.setData(allData);
                    }
                }
            }
            
            // Display trades table
            if (data.trades && data.trades.length > 0) {
                const tableDiv = document.getElementById('tradesTable');
                let tableHTML = '<h3>Trade History</h3><table><thead><tr><th>ID</th><th>Time</th><th>Price</th><th>Quantity</th><th>Side</th><th>Commission</th><th>Profit/Loss</th></tr></thead><tbody>';
                
                // Calculate profit/loss for each trade
                let previousTrade = null;
                data.trades.forEach(trade => {
                    const tradeTime = new Date(trade.time).toLocaleString();
                    let profitLoss = '';
                    
                    // Calculate profit/loss if we have a previous trade
                    if (previousTrade) {
                        if (previousTrade.is_buy && !trade.is_buy) {
                            // Buy then sell - calculate profit
                            const buyValue = previousTrade.price * previousTrade.qty;
                            const sellValue = trade.price * trade.qty;
                            const tradeProfit = sellValue - buyValue;
                            profitLoss = '$' + tradeProfit.toFixed(2);
                        } else if (!previousTrade.is_buy && trade.is_buy) {
                            // Sell then buy - calculate profit
                            const sellValue = previousTrade.price * previousTrade.qty;
                            const buyValue = trade.price * trade.qty;
                            const tradeProfit = sellValue - buyValue;
                            profitLoss = '$' + tradeProfit.toFixed(2);
                        }
                    }
                    
                    tableHTML += '<tr><td>' + trade.id + '</td><td>' + tradeTime + '</td><td>$' + trade.price.toFixed(2) + '</td><td>' + trade.qty + '</td><td>' + (trade.is_buy ? 'BUY' : 'SELL') + '</td><td>$' + trade.commission.toFixed(4) + '</td><td>' + profitLoss + '</td></tr>';
                    previousTrade = trade;
                });
                
                tableHTML += '</tbody></table>';
                
                tableDiv.innerHTML = tableHTML;
            }
        }
        
        // Handle window resize
        window.addEventListener('resize', () => {
            if (pricePlot) {
                pricePlot.setSize(document.getElementById('priceChart').offsetWidth, 500);
            }
            if (equityPlot) {
                equityPlot.setSize(document.getElementById('equityChart').offsetWidth, 500);
            }
        });
    </script>
</body>
</html>`
}
