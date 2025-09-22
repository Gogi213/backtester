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
        paramsDiv.innerHTML = '<label for="bbPeriod">Period:</label><input type="number" id="bbPeriod" value="10" min="2"><label for="bbStdDev">Standard Deviations:</label><input type="number" id="bbStdDev" value="1" step="0.1" min="0.1">';
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
    
    metricsDiv.innerHTML = `
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
    `;
    
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
                        [100, "{HH}:{mm}:{ss}", null, 1]
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
        let tableHTML = '<h3>Trade History</h3><table><thead><tr><th>Entry Time</th><th>Entry Price</th><th>Exit Time</th><th>Exit Price</th><th>Side</th><th>Quantity</th><th>Commission</th><th>Profit/Loss</th></tr></thead><tbody>';
        
        // Group trades into entries and exits
        for (let i = 0; i < data.trades.length; i += 2) {
            if (i + 1 < data.trades.length) {
                const entryTrade = data.trades[i];
                const exitTrade = data.trades[i + 1];
                
                const entryTime = new Date(entryTrade.time).toLocaleString();
                const exitTime = new Date(exitTrade.time).toLocaleString();
                
                // Calculate profit/loss
                let profitLoss = 0;
                let commission = entryTrade.commission + exitTrade.commission;
                
                if (entryTrade.is_buy) {
                    // Long position: profit = exit price - entry price
                    profitLoss = (exitTrade.price - entryTrade.price) * entryTrade.qty - commission;
                } else {
                    // Short position: profit = entry price - exit price
                    profitLoss = (entryTrade.price - exitTrade.price) * entryTrade.qty - commission;
                }
                
                tableHTML += '<tr><td>' + entryTime + '</td><td>$' + entryTrade.price.toFixed(2) + '</td><td>' + exitTime + '</td><td>$' + exitTrade.price.toFixed(2) + '</td><td>' + (entryTrade.is_buy ? 'LONG' : 'SHORT') + '</td><td>' + entryTrade.qty.toFixed(4) + '</td><td>$' + commission.toFixed(4) + '</td><td style="' + (profitLoss >= 0 ? 'color: green;' : 'color: red;') + '">$' + profitLoss.toFixed(2) + '</td></tr>';
            }
        }
        
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