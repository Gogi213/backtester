# План реализации графика цены с отображением сделок

## 1. Анализ текущей архитектуры

### Текущий поток данных:
1. Бэкенд загружает данные по часам через `LoadTradesByHour` в `handlers/handlers.go`
2. Данные передаются в `BacktestEngine.Run()` в `backtester/engine.go`
3. Результат бэктеста возвращается как `BacktestResult` с полями:
   - `PriceData` - данные для графика цены
   - `Trades` - массив сделок
   - `EquityCurve` - кривая капитала
4. На фронтенде отображается только equity curve через uPlot

### Структуры данных:
- `ChartPoint` (в `handlers/handlers.go` и `backtester/engine.go`) - точка графика цены
- `Trade` (в `backtester/backtester.go`) - сделка из бэктеста

## 2. Необходимые изменения

### 2.1. Добавить контейнер для графика цены
- Модифицировать HTML-шаблон в `templates/templates.go`
- Добавить контейнер `<div id="priceChart"></div>` перед equity chart

### 2.2. Модифицировать JavaScript
- Обновить `templates/script.js` для отображения графика цены
- Добавить функцию отображения графика цены с данными из `data.price_data`
- Добавить отображение сделок на графике как точки

### 2.3. Проверить передачу данных
- Убедиться, что `PriceData` передается в JSON-ответе бэктеста
- Проверить формат данных для графика

## 3. Подробная реализация

### 3.1. Изменения в HTML-шаблоне
```html
<!-- Добавить после заголовка, но перед equity chart -->
<div class="chart-container">
    <h3>Price Chart with Trades</h3>
    <div id="priceChart"></div>
</div>
```

### 3.2. Изменения в JavaScript
Добавить функцию отображения графика цены:
```javascript
function displayPriceChart(priceData, trades) {
    if (!priceData || priceData.length === 0) return;
    
    // Подготовить данные для графика
    const timestamps = priceData.map(point => point.time / 1000);
    const prices = priceData.map(point => point.price);
    
    // Подготовить данные сделок
    const buyTrades = trades.filter(trade => trade.is_buy);
    const sellTrades = trades.filter(trade => !trade.is_buy);
    
    const buyTimes = buyTrades.map(trade => new Date(trade.time).getTime() / 1000);
    const buyPrices = buyTrades.map(trade => trade.price);
    
    const sellTimes = sellTrades.map(trade => new Date(trade.time).getTime() / 1000);
    const sellPrices = sellTrades.map(trade => trade.price);
    
    // Опции графика
    const opts = {
        title: "Price Chart with Trades",
        id: "price-chart",
        class: "my-chart",
        width: document.getElementById('priceChart').offsetWidth,
        height: 500,
        series: [
            {},
            {
                label: "Price",
                stroke: "blue",
                width: 2,
                points: { show: false }
            },
            {
                label: "Buy Trades",
                stroke: "green",
                width: 0,
                points: { show: true, size: 5, fill: "green" }
            },
            {
                label: "Sell Trades",
                stroke: "red",
                width: 0,
                points: { show: true, size: 5, fill: "red" }
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
                label: "Price ($)",
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
    
    // Создать график
    if (window.pricePlot) {
        window.pricePlot.destroy();
    }
    
    window.pricePlot = new uPlot(opts, [timestamps, prices, buyTimes, buyPrices, sellTimes, sellPrices], document.getElementById('priceChart'));
}
```

### 3.3. Обновление функции displayResults
```javascript
function displayResults(data) {
    // ... существующий код ...
    
    // Отобразить график цены с сделками
    displayPriceChart(data.price_data, data.trades);
    
    // ... остальной существующий код ...
}
```

## 4. Проверка и тестирование

1. Убедиться, что данные передаются корректно
2. Проверить отображение графика цены
3. Проверить отображение сделок как точек
4. Убедиться, что график адаптируется к размеру окна
5. Протестировать с разными часовыми периодами

## 5. Возможные проблемы и решения

1. **Проблема**: Большие объемы данных могут замедлить отображение
   **Решение**: Добавить сэмплирование данных на бэкенде при необходимости

2. **Проблема**: Наложение точек сделок на график цены
   **Решение**: Использовать разные серии данных в uPlot для корректного отображения

3. **Проблема**: Неправильное форматирование времени
   **Решение**: Убедиться, что все временные метки в миллисекундах и правильно конвертируются