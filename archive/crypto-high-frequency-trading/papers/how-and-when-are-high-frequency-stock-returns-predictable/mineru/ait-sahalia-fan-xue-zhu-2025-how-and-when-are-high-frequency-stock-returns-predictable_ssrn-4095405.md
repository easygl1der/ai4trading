# How and When are High-Frequency Stock Returns Predictable? ∗

Yacine A¨ıt-Sahalia<sup>†</sup>

Jianqing Fan<sup>‡</sup>

Lirong Xue

Department of Economics

Department of ORFE

Department of ORFE

Princeton University and NBER

Princeton University

Princeton University

Xiaonan Zhu

Department of ORFE

Princeton University

This Version: October 15, 2025

## Abstract

This paper studies the predictability of ultra-high-frequency stock returns and durations to relevant price, volume, and transaction events using machine learning methods. We find that contrary to low-frequency and long-horizon returns, where predictability is rare and inconsistent, predictability in high-frequency returns and durations is large, systematic, and pervasive over short horizons. We identify the relevant predictors constructed from trades and quotes data and examine what determines the variation in predictability across the stock’s own characteristics and market environments. Next, we compute how the predictability improves with the timeliness of the data on a scale of milliseconds, and conversely degrades with delays, providing a valuation of each millisecond gained. Finally, we simulate the impact of getting an (imperfect) peek at the incoming order flow, a look-ahead ability that is often attributed to the fastest high-frequency traders, in terms of improving the predictability of the returns and durations.

Keywords: High-frequency returns; Durations; Predictability; Machine learning; Random forests; LASSO; Penalized Regression.

JEL Codes: G12; G14; C45; C53; C58.

## 1. Introduction

Low-frequency predictability of asset returns, for example over months or years, has been extensively studied and hotly debated in the literature: classical examples of the two sides of the debate are Fama (1970) and Malkiel (1973) vs. Lo and MacKinlay (2002). To the extent that such predictability is present in the data, the empirical evidence suggests that it is overall relatively small and dificult to pin down, and depends heavily upon the stocks or sectors studied, the predictor variables included, the horizon and time periods considered as well as the methodology employed. Given low signal-to-noise ratios, weak and persistent predictors, and instability of the predictive relations (see, e.g., Timmermann (2018)), it is not surprising that achieving consistent low frequency returns predictability is challenging.

By contrast, we show in this paper that the predictability of returns and durations is systematic and pervasive at high frequency over ultra-short horizons. Empirically, we find that almost every quantity of interest is strongly and consistently predictable over ultra-short horizons, across all stocks and periods in our sample. Specifically, price moves over short periods of time, their direction, magnitude, and momentum, and the duration between successive arrivals of both orders and transactions can be predicted. We show this using machine learning algorithms designed to extract predictability from high-dimensional data.

The question of predictability at low frequency has theoretical implications for whether financial markets are informationally eficient, as well as practical implications for asset allocation strategies. Predictability at high frequency, on the other hand, has theoretical implications for the design, operation, and regulation of financial markets, as well as practical implications for trading and execution strategies. Low-frequency predictability, if and when it is identified, should not be expected to last due to competitive pressures in the asset management industry and investors’ learning. By contrast, the technological costs and barriers to entry into high-frequency trading and the short shelf life inherent in ultra short horizon predictability make it more likely to persist.

Indirect evidence in favor of the presence of high-frequency predictability comes in the form of growth, and large and consistent profitability,<sup>1</sup> of some high frequency trading firms over the past fifteen years. It is implausible that such extraordinary profitability can be achieved simply by collecting compensation for intermediation services or trading profits, or as a reward for speed alone, in the absence of some ability to predict the short term direction of the order flow, price movements, or both (see Baron et al. (2019) and A¨ıt-Sahalia and Brunetti (2020)).

On the other hand, direct evidence in the literature regarding the extent, scope, and pervasiveness of high frequency predictability is less developed than at low frequency. Huang and Stoll (1994) use a two-equation econometric model of quote revisions and transaction returns to make short horizon returns predictions. Alvim et al. (2010) shows that the daily trading volume is predictable using high-frequency data. Zheng et al. (2013), Kercheval and Zhang (2015), Tsantekidis et al. (2017) and Ntakaris et al. (2018) show that information obtained from the limit order book can predict the direction of the next price move and crossings. Panayi et al. (2018) show that liquidity demand throughout the trading day is predictable and far from uniformly distributed. Knoll et al. (2019) show that trading signals extracted from Twitter data can predict returns. Sirignano (2019) shows that the liquidity sitting deep into the order book can be predicted. Chinco et al. (2019) show that one-minute-ahead returns can be forecast using the entire cross-section of lagged returns and identified as predictors stocks with news about fundamentals.

Easley et al. (2021) predict microstructure-based measures of market quality such as bid-ask spreads, the sign of change in realized volatility, the sign of change in the Jarque-Bera statistic, the sign of change in kurtosis, the sign of change in absolute skewness and the sign of change in the sequential correlation. By contrast, we examine the extent to which direct aspects of the next transaction or group of transactions (such as their direction, size, and price) as well as the duration of the next event (such as the next price change, or the next volume traded, or the next number of transactions) can be predicted. So while Easley et al. (2021) are predicting various moments or statistics of the price or returns process, we are predicting realizations of the price and volume processes directly. Forecasts of price direction over very short periods of time are inputs into directional positions either for an aggressive trading strategy or for passive market making.<sup>2</sup> Duration variables are important inputs to order placement and cancellation decisions, especially for liquidity providers whose limit orders may be queued in the book and require an estimate of the time before they reach the front of the line. Nevertheless, while predictability is an essential input to many trading strategies, a rigorous mapping from predictability to profitability remains highly complex.<sup>3</sup>

Besides quantifying the predictability that is present in the data, and how fast it dissipates, we seek to determine which predictor variables are most informative for the next trade and durations. We use only widely available trade and quotes data. For both directional traders and market makers, one of the main source of data that is relevant for the decision to send in a trade order, place or withdraw quotes on these very short time scales – where fundamental information about the asset being traded is unlikely to have changed – are variables that can be exclusively inferred from recent trades, as well as the current state and recent evolution of the limit order book. By construction, the information set we base our predictions on is a subset of the information set available to market participants, who might conceivably be able in real time to extract further information from observing the asynchronously incoming order flow on a variety of linked exchanges where they may have posted orders. As a result, the amount of predictability we identify using merely transactions and quotes data, although already quite strong, is a lower bound on the amount that is in principle achievable by a fast and technologically up-to-date market participant.

The types of machine-leaning methods we employ in this context are diferent from probabilistic or statistical approaches that model the limit order book (see e.g., Cont et al. (2010), Ro¸su (2009)), which first often rely on specific functional form assumptions, and second are not primarily designed to produce out-of-sample forecasts. In any event, one of the findings in this paper is that variables derived from the state and evolution of the limit order book tend to be less useful for predicting future returns than variables derived from the transactions record. Among the machine learning methods we consider, the specific method employed generally makes little diference to the outcome provided the methods are trained to the same data and fine-tuned to a comparable degree.

More specifically, we study the predictability over ultra short horizons of transaction returns, directions, and durations to specific prices, volumes, or transaction events. Each problem is tested with diferent machine learning methods over diferent time horizons and clocks, including calendar, trade, and volume clocks. We use the complete transactions and quote data for the 100 stocks that were constituents of the S&P 100 index over the January 2019 - December 2020 period. In typical machine learning fashion, we construct large numbers of data features using microstructure and other measures, use various lags, as well as combinations thereof. We find that out-of-sample predictability exists universally in all the stocks examined, and all the time periods, including the highly volatile environment of March and April 2020. With minimal algorithm tuning, for the median stock in the sample, a 10.5% out-of-sample $R ^ { 2 }$ for predicting 5-second returns can be achieved using merely past trade and quote data and an accuracy of 64% for predicting the direction of the next trade. When predicting the duration till the next 10 trades, the median out-of-sample $R ^ { 2 }$ is 9.8%. For return and direction predictions, important predictors include the imbalance in the limit order book, recent transaction imbalance, and past trade returns, while statistics derived from recent trade volume are most efective for duration predictions.

Second, we examine how long high frequency predictability lasts for a typical stock in the sample. We find that the predictability of returns vanishes to 0 (meaning that a binary directional prediction becomes no better than a coin toss) in just five minutes, and approximately 2, 000 transactions or 2, 000 lots transacted.

Third, we quantify the importance of the timeliness of the data. This is a key practical consideration as some nonzero delay is unavoidable due to physical and computational constraints. We show that approximately 80% of the overall predictability is achieved by relying on the most recent 10 milliseconds, 10 transactions or 10 lots transacted. We also show that introducing a small delay in acquiring and processing the data, in the form of a lag, decreases sharply the accuracy of the predictions. We map the decline as a function of the delay. Using this method, we can assign a value to each millisecond of delay, and rationalize the vast investments and sometimes extreme steps undertaken by high frequency market participants to further lower their latency.

Fourth, we simulate the efect that acquiring some signal on the direction of the order flow would have for the accuracy of the predictions. The idea here is to model a high frequency trader who, through order placement in diferent exchanges, is able to (imperfectly) infer the direction of the next trade as suggested in, e.g., Lewis (2015). We ask how much such information would improve forecasts of the next price move, as a function of the accuracy of the directional signal. Such ability to “look ahead” at the incoming flow, even limited to an imperfect sign prediction, is able to boost 5-second return R<sup>2</sup> from 14.0% up to 27.1% and the price direction accuracy from 68.3% up to 79.0%. Whether it is realistic or not to endow some high frequency traders with such an ability, it is clearly valuable.

Fifth, we investigate whether any diferences in predictability can be explained by stock-level characteristics and the market environment. Determinants of predictability, including stocks’ own characteristics combined with cross-sectional and time-series factors, are studied using panel regressions. Returns and trade direction are found to be more predictable for stocks that have smaller nominal share prices, that are less liquid, less volatile, and less correlated with the aggregate market.<sup>4</sup> These aspects, together with fixed efects for dates and stocks, explain nearly three quarters of the variability of 5-second return R<sup>2</sup>s. By contrast, predictability for durations is higher under liquid and volatile conditions.

Finally, we check the robustness of the results by using diferent machine learning methods and tuning algorithms, determine how predictability varies at diferent times of day, the relative value of the two subtypes of data (trades and quotes), as well as the incremental value of supplementing the data on a given stock with data from other (correlated) stocks for the diferent prediction objectives.

Having described what this paper does, it is important to state what the paper does not do. This paper is about quantifying the predictability in high frequency short horizon prices, showing how it can be achieved in terms of methods and data, and understanding the impact of diferent environments. It is not about why such predictability is present in the first place. Answering the ’why’ question is certainly interesting, but the ’how and when’ analysis in this paper is a natural prerequisite. Understanding the underlying sources of predictability, from specific aspects of the design and operations of financial markets to modeling traders’ strategies, would require a very diferent set of methods than those employed in this paper.

The paper is organized as follows. We start in Section 2.1 by describing the diferent prediction problems and their corresponding response variables, the construction of the predictor variables, and the criteria we employ to evaluate model performance. Section 2.2 introduces the data and machine learning approaches we use. Section 3 presents the stock-level results. In Section 4, we study how the timeliness of the data and the ability to look ahead at the sign of the order flow change the predictability results. In Section 5, we turn to the question of determining what explains any diferences across stocks and time in the level of predictability we found: we examine diferent stock characteristics and market conditions, such as diferences in aggregate volatility. Section 6 reports the results of a series of robustness checks. Finally, Section 7 concludes this paper.

Table 1: Examples of trade data: INTC on Jan. 3rd, 2019

<table><tr><td>Time</td><td>Price</td><td>Size</td><td>Direction (Lee-Ready)</td></tr><tr><td>10:07:48.956770900</td><td>45.18</td><td>100</td><td>-1</td></tr><tr><td>10:07:48.956773554</td><td>45.18</td><td>300</td><td>-1</td></tr><tr><td>10:07:48.956916983</td><td>45.18</td><td>100</td><td>-1</td></tr><tr><td>10:07:48.956971093</td><td>45.18</td><td>100</td><td>+1</td></tr><tr><td>10:07:48.957830128</td><td>45.18</td><td>66</td><td>+1</td></tr></table>

## 2. Data and Methods

## 2.1 Response and Predictor Variables

The raw data take the form of the complete record of transactions and quote updates, along with their calendar timestamp. In the following, we define more precisely the variables of interest that we construct from the raw data, first the response variables and then the predictors.

## 2.1.1 Transactions and Quotes Data

We use only standard, widely-available data to construct our variables. Namely, all the variables are derived from the NYSE’s Trade and Quote (TAQ) database for two full years 2019 and 2020. TAQ contains consolidated intraday trades and level-1 quotes (best bids and ofers on the market) of all securities listed on the New York Stock Exchange (NYSE), NASDAQ, and American Stock Exchange (AMEX). We restrict attention to the 100 stocks that were constituents of the S&P 100 index on December 31, 2020.<sup>5</sup> Although heavily weighted towards more liquid stocks compared to the full stock universe, this provides a wide sample of stocks across all industry groups. A summary of the size of the dataset we use is shown in Table 6 of the Supplemental Material.

Table 2: Example of quote data: INTC on Jan. 3rd, 2019

<table><tr><td>Time</td><td>Best Bid Price</td><td>Best Bid Size</td><td>Best Ask Price</td><td>Best Ask Size</td></tr><tr><td>10:07:48.956906761</td><td>45.18</td><td>100</td><td>45.19</td><td>4800</td></tr><tr><td>10:07:48.956921135</td><td>45.18</td><td>100</td><td>45.19</td><td>4700</td></tr><tr><td>10:07:48.956970663</td><td>45.17</td><td>1600</td><td>45.19</td><td>4700</td></tr><tr><td>10:07:48.956980355</td><td>45.17</td><td>1600</td><td>45.19</td><td>4100</td></tr><tr><td>10:07:48.956991775</td><td>45.17</td><td>1600</td><td>45.19</td><td>4000</td></tr></table>

Tables 1 shows an example of transaction data for Intel Corporation. For each given date and ticker symbol, a row in the transactions data reports a single transaction. It contains a timestamp of the transaction and its associated price, size, and trading direction. For all the stocks we consider, quote and trade prices are reported as multiples of \$0.01.<sup>6</sup> The timestamp is measured in nanoseconds. We follow the usual Lee and Ready (1991) algorithm to infer order direction from the sequence of the trades, and indicate the trade direction as +1 if it is a buy-initiated trade and −1 if it is a sell-initiated trade. <sup>7</sup>

A snapshot of the quote update data is illustrated in Table 2. Each row in the quote data corresponds to the NBBO at a certain timestamp. The third line of quote update is likely caused by the fourth transaction shown in Table 1. O’Hara et al. (2014) report possible issues with the lack of records of odd-lot trades when TAQ only recorded round-lot trades; TAQ started to include odd-lot trades since 2014, as we can see in Table 2. The quotes are still round-lot, but this should have only a minimal impact on our response variables.

The transaction and quote update data are then merged based on their timestamps. The best bid and ask information of trades are determined by the most recent quote information as of the time of the transaction. We remove all observations outside normal trading hours so every timestamp is within 9:30:00 and 16:00:00 (strictly). Both transaction data and quote update data are used in tuning, training, and testing of the algorithms. Summary statistics of our data and response variables are reported in Table 7 in the Supplemental Material.

## 2.1.2 Response Variables

The response or dependent variables we investigate are transaction returns and direction, and the duration of relevant market events, such as the the arrival of the next group of trades, the next volume quantity, or the time to the next price change.

Time clocks. Duration quantities can be measured using three diferent time clocks: a calendar clock (corresponding to the usual measurement of time), a transaction clock (where time is measured in terms of the number of transactions that have taken place, irrespectively of their size) and a volume clock (where time is measured in terms of the total volume transacted, irrespectively of the number of transactions). We examine and contrast predictability in all three clocks.

The variables that follow are all defined for each stock individually. All the trades and quotes for a given stock are uniquely labeled by the timestamp of their occurrence in calendar time $t \in \mathbb { R } ^ { + }$ . Here t represents the number of seconds since the beginning of the day, with precision up to a nanosecond $( 1 0 ^ { - 9 } )$ . For example, $t = { 3 5 4 9 3 . 1 3 2 2 7 3 8 7 3 }$ represents calendar time 9:51:33.132273873. We denote the number of shares traded at t as $V _ { t }$ , with $V _ { t } = 0$ indicating no trade at t. Thus, $\mathbb { 1 } _ { \{ V _ { t } > 0 \} }$ indicates whether a transaction has taken place at time t. The calendar time interval between two timestamps $T _ { 1 } , T _ { 2 } \in \mathbb { R }$ can be defined as the following half open half closed interval

$$
\operatorname{Int} \left(T _ {1}, T _ {2}\right) = \left\{t \in \mathbb {R}: T _ {1} <   t \leq T _ {2} \right\}.\tag{2.1}
$$

For convenience and brevity of exposition, we extend the notion of interval to work for any of the three time clocks, in order to use a unified interval definition when considering each response variable. With a starting timestamp $T .$ , a span $\Delta > 0$ and a clock mode $M \in$ {calendar, transaction, volume}, we define the forward looking time interval as

$$
\operatorname{Int} ^ {\text {forward}} (T, \Delta , M) = \left\{ \begin{array}{l l} \operatorname{Int} (T, T + \Delta) & \text {if} M = \text {calendar} \\ \left\{t > T: \left(\sum_ {s \in \operatorname{Int} (T, t)} \mathbb {1} _ {\{V _ {s} > 0 \}}\right) \leq \Delta \right\} & \text {if} M = \text {transaction}. \\ \left\{t > T: \left(\sum_ {s \in \operatorname{Int} (T, t)} V _ {s}\right) \leq \Delta \right\} & \text {if} M = \text {volume} \end{array} \right.\tag{2.2}
$$

Under the calendar clock, $\Delta$ measures the target time horizon over which the prediction takes place. Under the transaction clock, $\Delta$ measures the target number of transactions: the interval contains all the timestamps after $T$ such that the total number of trades after $T$ is no more than $\Delta$ Under the volume clock, $\Delta$ measures the target volume: the interval contains all the time stamps such that the total volume traded after $T$ is no larger than $\Delta .$ . The interval is defined as a set of consecutive timestamps and will be used to define relevant quantities such as the average return and duration over that interval. The interval does not contain the starting timestamp $T .$

Let $\mathbf { D } ^ { \mathrm { t x n } }$ denote the set of all record timestamps $t \in \mathbb { R } ^ { + }$ corresponding to trade transactions and $\mathbf { D } ^ { \mathrm { q t } }$ its quote counterpart. Set $\mathbf { D } = \mathbf { D } ^ { \mathrm { t x n } } \cup \mathbf { D } ^ { \mathrm { q t } }$ . The National Best Bid and Ofer (NBBO) prices are indexed by $t \in \mathbf { D }$ and denoted as $\left( P _ { t } ^ { b } , P _ { t } ^ { a } \right)$ where $P _ { t } ^ { b }$ is the best bid price and $P _ { t } ^ { a }$ the best ask price. The mid-price is their simple average

$$
P _ {t} = \frac {P _ {t} ^ {b} + P _ {t} ^ {a}}{2}.\tag{2.3}
$$

Denote by $P _ { t } ^ { \mathrm { t x n } }$ the transacted price if $t \in \mathbf { D } ^ { \mathrm { t x n } }$ . The best bid and ask sizes are denoted $S _ { t } ^ { b }$ and $S _ { t } ^ { a }$ respectively for the record indexed by t.

Transaction return. At time $T \in \mathbf { D }$ , with span $\Delta$ and clock mode $M ,$ , the transaction return is defined as

$$
\operatorname{Return} (T, \Delta , M) = \operatorname{Average} \left[ P _ {t} ^ {\mathrm{txn}}: t \in \mathbf {D} ^ {\mathrm{txn}} \cap \operatorname{Int} ^ {\mathrm{forward}} (T, \Delta , \mathrm{M}) \right] / P _ {T} - 1.\tag{2.4}
$$

This quantity measures the average return from transactions in a forward window, typically over a very short horizon (∆ small). Compared with the return from a single transaction or at a fixed point in time, this definition leads to a less noisy response variable and is more indicative of the aggregated trading behavior over the span $\Delta$ instead of the information at an arbitrary point in the near future. From the perspective of a market maker with quotes in the limit order book, and deciding whether to keep or cancel them, predicting (2.4) is more relevant since, unless the quotes are at the front of the queue, the market maker cannot be certain of the exact execution timeframe, but can reasonably infer that it will take place over a short horizon depending upon the current state of the book and market activity. 8

Price direction. Next, we are interested in predicting whether the next (set of) price movements will be up or down. At time $T \in \mathbf { D }$ , with horizon $\Delta$ and time clock $M$ , the trade direction is defined as

$$
\operatorname{Direction} (T, \Delta , M) = \mathbb {1} _ {\{\operatorname{Return} (T, \Delta , M) > 0 \}}.\tag{2.5}
$$

This variable further normalizes the transaction return by making it a binary variable, regularizing outliers and tail behavior, and potentially facilitating the prediction, but at the cost of losing the information contained in the magnitude of the return. Nevertheless, direction is by itself a very important variable, and an accurate directional prediction can be one of the most basic inputs to a trading algorithm.

Transaction duration. At time $T \in \mathbf { D }$ , with span $\Delta$ and clock $M \in \{ \mathrm { t r a n s a c t i o n } , \mathrm { v o l u m e } \}$ , we define the duration variable as:

$$
\mathrm{Duration} (T, \Delta , M) = \operatorname * {a r g m a x} _ {t \in {\bf D}} \left\{t \in \mathrm{Int} ^ {\mathrm{forward}} (T, \Delta , \mathrm{M}) \right\} - T.\tag{2.6}
$$

This measures the amount of time it takes to record either $\Delta$ transactions or $\Delta$ shares: by definition, Duration $( T , \Delta , \mathrm { c a l e n d a r } ) = \Delta$ so the measurement is only relevant when M is either transaction or volume.<sup>9</sup>

The transaction duration variable measures trading intensity; predicting it is an input for order execution strategies as well as quote placement or cancellation strategies. For example, a trader seeking to place an order for a quantity that is not small relative to $\Delta$ might place a partial order and keep the rest in reserve. Predicting duration could help to set the rest of that trader’s execution schedule. Similarly, market makers who do not wish to see their quotes hit (for inventory or other reasons) might want to cancel them before they rise near the front of the queue; predicting duration helps determine how fast quotes need to be canceled.

## 2.1.3 Predictor Variables

We consider a large number of predictor variables representing the short-term trading environment for a particular stock. These variables are computed on nine disjoint windows, covering the most recent records (on a scale of 0.1 seconds) to a longer time horizon (on a scale of 25 seconds) so as not to prejudge the results in favor of a short lookback window. Specifically, we extracted 15 features from the TAQ data and computed these variables at nine look-back disjoint windows. For the calendar clock, nine look-back windows are used: $\{ ( 0 , . 1 ) , \ ( . 1 , . 2 ) , \ ( . 2 , . 4 ) , \ . \ . . , ( 1 2 . 8 , 2 5 . 6 ) \}$ with number of seconds as the unit; for the transaction clock, the 9 spans are $\{ ( 0 , 1 ) , ( 1 , 2 ) , ( 2 , 4 ) , \ldots , ( 1 2 8 , 2 5 6 ) \}$ using the number of transactions as a unit; for the volume clock, we use the number of shares traded: {(0, 100), (100, 200), (200, 400), . . . , (12800, 25600)}.

The predictors that we create include features on volume and duration: Breadth, Immediacy, VolumeAll, VolumeAvg, and VolumeMax. We also use the predictors derived from the return and imbalance of the limit order book. They are Lambda, LobImbalance, TxnImbalance, and PastReturn. In addition, we employ the predictors that measure the speed and cost in stock’s trading: Turnover, AutoCov, QuotedSpread, EffectiveSpread, RealizedVolatility, and TSRV. Details can be found in §B of the Supplemental Material.

## 2.2 Machine Leaning Methods

## 2.2.1 Models

The two main methods that we use to predict stock returns and durations are the regularized or penalized linear regression (in the form of least absolute shrinkage and selection operator or LASSO) as a representative parametric method, and random forests (RF) as a representative nonparametric one. These two methods are described in §G of the Supplemental Material. In Section 6.1 below, we perform a horse race across a large number of methods, including the two already mentioned and ordinary least squares (OLS), ridge regression, FarmPredict linear regression, and gradient boosted trees (GBT). More details containing these various methods can be found in Hastie et al. (2009) and Fan et al. (2020b).

## 2.2.2 Measuring Prediction Accuracy

We now describe how we measure and compare prediction results produced by diferent methods. For robustness, we use diferent criteria for the same prediction problem, emphasizing diferent aspects of the quality of the prediction made. For the prediction of stock returns, we employ both the out-ofsample coeficient of determination $R ^ { 2 }$ and the directional accuracy as measures of fit. $R ^ { 2 }$ measures how precisely targets are predicted in a normalized fashion, while the directional accuracy focuses on whether the predicted values are on the correct side of the target. Both metrics are applied to the testing dataset. Typically, $R ^ { 2 }$ will be used when assessing the performance of trade returns and duration predictions, as well as in tuning model hyper-parameters, while accuracy will be used for evaluating trade direction predictions.

Out-of-sample $R ^ { 2 }$ is a widely used criterion for evaluating regression models. It measures the normalized prediction error in the target Y in comparison with a trivial prediction, such as the sample average. For given targets Y and predictions $\widehat { \mathbf Y }$ , the out-of-sample $R ^ { 2 }$ is defined as

$$
R ^ {2} (\mathbf {Y}, \widehat {\mathbf {Y}}) = 1 - \frac {\sum_ {i} (Y _ {i} - \widehat {Y _ {i}}) ^ {2}}{\sum_ {i} (Y _ {i} - \bar {Y}) ^ {2}},\tag{2.7}
$$

where $\bar { Y }$ is the in-sample average calculated based on the training set. The measure produces a value $R ^ { 2 } \in ( - \infty , 1 ]$ , and the larger the score the better the model performance. When $R ^ { 2 } > 0$ , the corresponding method outperforms the sample average estimator. In the limit, $R ^ { 2 } = 1$ means that the target can be perfectly predicted.

Note that out-of-sample $R ^ { 2 }$ aggregates the squared errors for testing data with the same weight, which can be influenced by outliers of the prediction errors. Unfortunately, prediction errors follow heavy-tailed distributions, as stock prices jump frequently, creating large outliers in returns. Similarly, for example, the trading volume varies substantially over time, with disproportionately large orders from time to time, resulting in outliers in the duration variables. These considerations also lead us to consider a more robust measurement, the sign accuracy, that is less sensitive to outliers.

We use sign accuracy to measure the results of directional predictions. Let Y be the target and $\widehat { \mathbf Y }$ be the prediction, and consider the accuracy measure:

$$
\mathrm{Accuracy} (\mathbf {Y}, \widehat {\mathbf {Y}}) = \frac {1}{n} \sum_ {i} \mathbb {1} _ {\left\{\widehat {Y _ {i}} \cdot Y _ {i} > 0 \right\}}.\tag{2.8}
$$

This measure calculates the proportion of predicted $\widehat { Y } _ { i }$ that have the same sign (or direction) as the true $Y _ { i }$

Since $R ^ { 2 }$ is not robust to outliers and the direction accuracy does not take into account the magnitude of the returns, we add a more balanced measure of accuracy by introducing a finer grid of classification buckets in order to maintain some robustness to outliers while retaining some ability to condition on the returns size. Specifically, we divide returns Y into quartile buckets based on the testing data. Once we make predictions $\widehat { \mathbf Y }$ , we obtain a $4 \times 4$ confusion matrix for $\widehat { \mathbf Y }$ with respect to the true classes generated based on Y by counting how many times a prediction is made in a given quartile given the quartile that contained the testing return. Counts on the diagonal of the confusion matrix correspond to return predictions that are correct, in the sense that they have been made in the same quartile as the testing return. So we define the prediction quartile accuracy as

$$
\mathrm {Quartile\_Accuracy} (\mathbf {Y}, \widehat {\mathbf {Y}}) = \frac {1}{n} \sum_ {i} \mathbb {1} _ {\left\{\text {class of} \widehat {Y} _ {i} = \text {class of} Y _ {i} \right\}},\tag{2.9}
$$

where Y and $\widehat { \mathbf Y }$ are the target and prediction, and class of $Y _ { i } \ ( \mathrm { r e s p . } \ \widehat { Y } _ { i } )$ is the intervals that $Y _ { i }$ (resp. $\widehat { Y } _ { i } )$ lies in with respect to the four quartiles of Y.

## 2.2.3 Algorithm Tuning and Testing

We employ a rolling window approach to ensure that models are continuously updated with the most recent data. Specifically, on each testing day, a model is trained using data from the previous five trading days, while hyperparameters are re-tuned every 20 trading days. Detailed implementation procedures are provided in §G.3 of the supplemental material.

Figure 1: Distribution of average out-of-sample $R ^ { 2 }$ when predicting individual stock returns  
![](images/0740a13aebe21b3b6e6c6dad0b56d616f244866d764e7128d20e01920b99c9bf.jpg)  
Note: The x-axis shows the look-forward horizon of each return prediction in three diferent clocks. Each box plot summarizes the distribution of the average out-of-sample $R ^ { 2 } { \mathrm { : } }$ , which measures the overall daily performance of a security in 2019 and 2020, across 100 stocks. The dashed line represents the average performance and the solid bar in the box indicates the median performance. The black horizontal line (marked baseline) corresponds to using the sample mean of the training data.

## 3. Predictability Results for Individual Stocks

In this section, we report the predictability results for all S&P 100 stocks over the 505 trading days in the 2019-2020 period. As described in Section 2.1.3, we use 15 variables defined there over 9 diferent time windows, resulting in 135 variables, that the learning algorithms are then free to further combine and interact as they see fit. The variables that have important predictive power for specific objectives are identified by LASSO and RF separately.

## 3.1 Transaction Return Prediction

We start by examining the ability to predict future stock returns over horizons of 5 and 30 seconds in calendar time, 10 and 200 transactions and 1,000 and 20,000 shares, using LASSO and RF, respectively. The prediction performance for individual stocks, as measured by out-of-sample $R ^ { 2 }$ averaged over the 505 days, is reported in Figure 1 in the form of boxplots summarizing the distribution of $R ^ { 2 }$ over the S&P 100 stocks. Recall from (2.7) that $R ^ { 2 } = 0$ is the baseline where predictability using an algorithm is no better than using the sample short horizon average return to predict the future return.

The main result here is that the algorithms, even with the limited tuning described above, largely outperform the baseline. The median out-of-sample $R ^ { 2 }$ for predicting 5-second returns is approximately 10% and furthermore $R ^ { 2 } > 0$ for every single stock, including the least predictable. The prediction results are slightly better using RF than LASSO. In addition, as expected, 30-second returns are harder to predict than 5-second ones, with a median $R ^ { 2 }$ of approximately 4%, but the returns of every single stock remain predictable. The results in trade and volume clocks are similar and consistent: strong predictability over the shorter horizon that gets weaker as the horizon increases. See Section §C for the results on the average daily performance across 100 stocks.

Figure 2: Top 20 explanatory variables selected by LASSO and RF for predicting 5-second returns  
![](images/204f24d68fbd4ebcd7d2a4d4bcd5604b6688929dec72f4a5e889cf84bb0e6de3.jpg)

![](images/d4b249f45ab977e63fcb1927f463dcd0023fb973e5b7a0ab7f3e110bb8607a7d.jpg)  
Coeficient magnitude for standardized predictors (LASSO)

![](images/cac8f2ba68e6ec74f7a4a97ff0e1d4e2b3b18e2861fe906e181e0317febbccbc.jpg)  
Mean Decrease in Impurity (MDI) for predictors (RF)  
Note: The top panel describes the frequency of mostly used predictor variables by LASSO. A variable is marked as used if its regression coeficient from LASSO is not zero. Frequencies are calculated over each test of 505 days and over 100 securities. The middle panel shows the average of the median coeficients of LASSO across all test with error bars indicating their 95% confidence intervals. Coeficients with the largest 5 absolute average values are shown. The y-axis shows the variables selected. The lower panel is similar to the middle one, showing the average of Mean Decrease in Impurity (MDI) of RF across all tests with error bars indicating their 95% confidence intervals. The values in brackets define a past interval. For example, (., 0.8, 1.6, calendar) includes all the data from the past 1.6 seconds to the past 0.8 seconds in the calendar clock

## 3.2 Variable Importance for Return Prediction

We next seek to determine which variables are most responsible for this predictability. Using LASSO, we measure both the frequency with which a variable is being selected (across stocks and days) and the size of the average regression coeficients over all LASSO regressions across stocks and days (recall that variables are standardized before LASSO is applied).

Using RF, we separately measure variable importance using the Mean Decrease in Impurity (MDI) for each variable, which is computed as follows. For each internal node of each decision tree, one feature is selected to make a decision on how to divide the data set into two separate sets with similar responses within. The feature is selected based on variance reduction from the regressions - the one with the highest decrease is set as the criterion for the corresponding internal node. The importance of a feature is computed as the (normalized) total reduction in variance brought by that feature over all trees in the forest, resulting in the MDI. A higher value indicates greater importance of the feature.

Figure 2 shows the results for predicting 5-second returns. Both methods give reasonably consistent results concerning the ranking of variables. The top two predictors are the transaction imbalance (TxnImbalance) and the past short horizon average return (PastReturn). Interestingly, both are derived from the transactions rather than the quotes data and capture some form of trade momentum. Both coeficients have the expected sign: for example, if there are many buy trades happening in a short look-back window, this trend is likely to persist for a short while and push the price upward. The next source of signal from the LASSO ranking comes from the quote data, in the form of the imbalance between the bid and ask side in the limit order book (LobImbalance). The coeficient also has the expected sign, namely that a bid size dominating the ask size predicts upward pressure on the price.

Another interesting observation is that the most informative predictors are constructed by using the most recent past data. The strongest signals always come from the most recent window, as we can see from the magnitude of the coeficients for the top three predictors, which are constructed over the most recent 0.1 seconds. Below, we will explore how fast the predictability deteriorates when the algorithms are precluded from exploiting the most recent past data, which simulates trader’s technological latency, including delay in receiving data and computation time.

The consistency of the selection of each variable is examined in Figure 3, where we plot the frequency with which a given variable is selected by the LASSO model, for the three time clocks and two time spans. Here, we aggregate the presence of a variable in the LASSO model at the group level (among a variable measured in 9 diferent time windows, as long as one of them is in the

Figure 3: Frequency of variable groups selected by LASSO for predicting returns over diferent horizons and time clocks  
![](images/660e6daaca575b6e11290fb678ee76339b47b59e7f0ab8533c50dd9e28f5a2c8.jpg)  
Note: Each variable in x-axis is measured in 9 diferent time windows. A group of variables is counted as selected if at least one of the variables in the group has non-zero LASSO coeficient. For a given past interval, Breadth measures the amount of transactions and Immediacy indicates the average interval between transactions. Vol umeAll, VolumeAvg, and VolumeMax are related to the total, average, and maximum traded volume. Lambda shows the price change in the interval proportional to total volume. LobImbalance and TxnImbalance are the limi order book imbalance and transaction imbalance for the interval. Turnover is the turnover rate and AutoCov is the auto-covariances of returns between consecutive transactions. EfectiveSpread is the dollar weighted spread on transactions. Detailed definitions of each variable can be found in Section 2.1.3 and Section §B of the Supplemental Material.

LASSO model, that group of variables is counted). The results reveal that variables in the group TxnImbalance, PastReturn, LobImbalance are almost always used, whereas the variables VolumeAll, VolumeAvg, VolumeMax are consistently not predictive in diferent models.

## 3.3 Price Direction Prediction

We now study the predictability of the return direction: is the price going up or down over the very short term? This binary classification ignores the magnitude of the return and counts only its sign, using the same LASSO and RF trained models. For reasons already discussed, this assessment of the predictions is less sensitive to outliers. Figure 4 shows the results in the form of the percentage of time the algorithm predicts the sign of the upcoming return correctly. The baseline from a random guess is 50%. The prediction accuracy is approximately 64% for returns over short horizons – the next 5 seconds, 10 trades or 1,000 shares transacted. Similar to the patterns in return predictions, the accuracy decreases as the time horizon increases. However, compared to Figure 1, there are some diferences shown in the figure. First, LASSO has substantially the same performance as RF. One reason is that RF is more robust than LASSO in terms of handling noisier data. The price direction response is less noisy than the returns, so $\mathrm { R F } ' \mathrm { s }$ advantage is less pronounced. Second, the result is more robust (i.e., less dispersed) than for predicting returns as can be seen from the lower interquartile range and fewer scattered outliers in the boxplots. The variables identified as important for direction prediction are similar to those for return prediction.

Figure 4: Distribution of average out-of-sample directional accuracy when predicting individual stock returns  
![](images/a3b6d92e763add3677c44e696ea2b265aed11d4281026aa8ed3e467e8f0ee88a.jpg)  
Note: The x-axis shows the look-forward time horizons of returns whose directions are to be predicted. Each box plot is summarized over 100 data points where each one is the average daily performance of a stock aggregated over 505 days in 2019 and 2020. The dashed line represents the average performance and the black horizontal line (baseline) indicates the result from using a random guess.

To show more precisely how the machine learning models perform in price direction prediction while including some information on the magnitude, Table 3 shows the 4×4 scaled confusion matrices for predicting quartiles of returns over the next 5 seconds and 30 seconds respectively with RF, as discussed in Section 2.2.2. The confusion matrices show that, among those predictions with correct signs, more than half of them correctly estimate the magnitude range. To see that, for example, in the confusion matrix for the 5 seconds return prediction, the sum of (2, 1)-th and (1, 2)-th elements is $0 . 1 5 8 + 0 . 0 1 0 = 0 . 1 4 8$ , and the sum of the (1, 1)-th and (2, 2)-th elements is $0 . 0 3 7 + 0 . 1 1 8 = 0 . 1 5 5$ meaning that 51.2% of the predictions that correctly demonstrate the positive direction also achieves the correct quartile buckets. Also, since the (2, 1)-the element is way greater than the (1, 2)-th element, we see that the prediction magnitude is more likely to be under-estimated than overestimated.

Table 3: Average scaled confusion matrix for quartile accuracy when predicting stock returns

<table><tr><td>Prediction</td><td colspan="4">Scaled Confusion Matrix</td><td>Quartile Accuracy</td></tr><tr><td rowspan="4">5s return</td><td>0.037 (0.002)</td><td>0.010 (0.000)</td><td>0.006 (0.000)</td><td>0.003 (0.000)</td><td rowspan="4">0.3387</td></tr><tr><td>0.158 (0.001)</td><td>0.118 (0.002)</td><td>0.114 (0.001)</td><td>0.071 (0.001)</td></tr><tr><td>0.072 (0.001)</td><td>0.087 (0.001)</td><td>0.149 (0.001)</td><td>0.141 (0.001)</td></tr><tr><td>0.003 (0.000)</td><td>0.004 (0.000)</td><td>0.011 (0.001)</td><td>0.035 (0.002)</td></tr><tr><td rowspan="4">30s return</td><td>0.023 (0.001)</td><td>0.007 (0.000)</td><td>0.004 (0.000)</td><td>0.003 (0.000)</td><td rowspan="4">0.3115</td></tr><tr><td>0.142 (0.001)</td><td>0.126 (0.001)</td><td>0.115 (0.001)</td><td>0.083 (0.001)</td></tr><tr><td>0.083 (0.001)</td><td>0.099 (0.001)</td><td>0.141 (0.001)</td><td>0.143 (0.001)</td></tr><tr><td>0.002 (0.000)</td><td>0.003 (0.000)</td><td>0.007 (0.000)</td><td>0.022 (0.001)</td></tr></table>

Note: The $4 \times 4$ scaled confusion matrix for $\widehat { \mathbf Y }$ with respect to the true quartile buckets generated based on Y. The two matrices are, respectively, for the prediction of future stock returns over horizons of 5 seconds and 30 seconds in calendar time. The ratio is averaged over all stocks. The brackets report standard errors of the mean across stocks. The confusion matrix is scaled such that its elements sum up to one, and the $( i , j ) \ – \mathrm { t h }$ element (for $i , j \in \{ 1 , 2 , 3 , 4 \} )$ is the percentage of return predictions that the prediction falls in the i-th bucket while the truth is in the j-th bucket.

Overall, the conclusion so far is that stock returns over short horizons are highly predictable using either LASSO or RF. Given the ultra short horizons over which it applies, such consistent predictability can, in principle, translate into profitability since a high frequency trader could trade thousands of times a day, replicated over a whole array of stocks, while predicting correctly the direction of the price change over 60% of the time. By the law of large numbers and central limit theorem, as the number of such trades increases, the fast trader is likely to come increasingly close to achieving this success rate in reality, potentially resulting in consistent profitability such as that reported in footnote 1. This said, drawing a precise mapping from a given amount of predictability to a given dollar amount of profitability is not straightforward, as already discussed in footnote 3. But it is obvious that identifying where and how predictability occurs can only help in the design and implementation of many trading strategies.

Figure 5: Distribution of average out-of-sample $R ^ { 2 }$ when predicting individual stock durations  
![](images/fab60c04706d988e7a65a1175bf733ea577f225a95712a881cecf5b435c216a3.jpg)  
Note: The x-axis shows diferent trading durations. Each box plot summarizes the distribution of the average out-of-sample $R ^ { 2 }$ of duration predictions, over 505 days in 2019 and 2020, across 100 stocks. The dashed line represents the average performance and the solid bar in the box indicates the median performance. The black horizontal line (marked baseline) corresponds to using the sample mean of the training data.

## 3.4 Transaction Duration Prediction

We now examine the predictability of trading durations, specifically targeting the amount of time necessary for a certain number of transactions to take place, or a certain volume to be traded. This variable is an important input into many execution strategies as well as price impact models. In the former application, optimally breaking a large transaction into smaller ones that are less conspicuous without incurring unnecessary delays during which the price could move adversely, requires an estimate of these durations. In the latter case, they give a time dimension to the price elasticity.

The methods we use for prediction are the same as those used for returns. Figure 5 reports the out-of-sample $R ^ { 2 }$ . Durations are predicted even more accurately than returns, and the predictability actually improves for longer durations (time to wait for 200 transactions to occur or 20,000 shares to change hands vs. time for 10 transactions or 1,000 shares), with once again $R ^ { 2 } > 0$ for every single stock. This reverse pattern compared to returns is due to the fact that, unlike returns over longer horizons, longer durations are more stable over time and hence easier to predict than shorter ones.

The predictor variables identified by either LASSO or RF as important for duration prediction are very diferent than those identified for predicting returns. The results are in Figure 6 (results using RF are similar and omitted to save space). The top 20 most significant features are all derived from two families of variables: VolumeMax and VolumeAll. The estimated regression coeficients of al variables of one family have the same signs, and the two families have opposite signs. The magnitude of the coeficient is proportional to the window length of the feature. A large value of VolumeAll indicates that trading activity is intense in the (very) recent past. This is likely to continue for a brief moment at least, as a result, the duration for a fixed number of shares is more likely to be predicted to be short, hence the negative value of the LASSO coeficients on VolumeAll. On the other hand, a large value of VolumeMax variable indicates that large size (perhaps block) trades are taking place at the moment. These large(r) trades may have more market impact but tend to be isolated and the LASSO model predicts that they are followed by longer than usual durations resulting in a positive coeficient for VolumeMax. So the two types of volume measurement play two very diferent roles in terms of forecasting durations. While the two volume variables are consistently the most important, the LASSO models for duration prediction use almost all other variables every time, so the grouped feature usage plots are omitted.

Figure 6: Top explanatory variables selected by LASSO when predicting duration  
![](images/3d224f19bc7d950eed28b3fcbcbbbc340e9bb2cb9a41bf9205104f636445c1b6.jpg)  
Note: All the explanatory variables are standardized before fitting LASSO and the variables with the largest 20 absolute LASSO coeficients are displayed in this plot. VolumeAll and VolumeMax measure the total traded volume and maximum single trade in a past interval. The values in brackets define the past interval. For example, (. , 32, 64, transaction) includes all the data after the past 64th transaction and before or at the past 32th transaction. The results for RF are omitted since they are similar.

## 3.5 Comparison with Microstructure Models’ Predictions

Some market microstructure models make specific parametric predictions regarding price changes that can be compared to those made by machine learning algorithms. In the model of Roll (1984), the eficient price $m _ { t }$ is a martingale, $m _ { t } = m _ { t - 1 } + u _ { t }$ where $u _ { t }$ are iid mean 0 and variance $\sigma _ { u } ^ { 2 } .$ and the transaction price is $p _ { t } = m _ { t } + c q _ { t }$ where $2 c$ is the bid-ask-spread and $q _ { t }$ is a trade direction indicator equal to +1 if the customer is a buyer and −1 if the customer is a seller. Buys and sells are equally likely and serially independent. The transaction price increment in this model has an MA(1) representation, $\Delta p _ { t } = \epsilon _ { t } + \theta \epsilon _ { t - 1 } , | \theta | < 1$ which can be fitted to the data and used to make predictions. Glosten and Harris (1988) extend the Roll model to allow for a volume efect: $m _ { t } = m _ { t - 1 } + u _ { t } , u _ { t } = w _ { t } + ( a _ { 0 } + a _ { 1 } V _ { t } ) q _ { t } , w _ { t }$ is iid mean 0, $p _ { t } = m _ { t } + ( c _ { 0 } + c _ { 1 } V _ { t } ) q _ { t }$ where $V _ { t }$ is the (unsigned) volume of the trade. In the ? model, the price change has the form $\Delta p _ { t } = \lambda ( x _ { t } + \epsilon _ { t } )$ where $x _ { t }$ is the informed trader demand, $\epsilon _ { t }$ is the uniformed volume, so the sum is the total order flow we observe. We regress the price change on the total order flow (signed) $V _ { t } = x _ { t } + \epsilon _ { t }$ , using the training set to estimate the parameter λ. In addition to Kyle’s original model, we also consider a square root price impact model (which is popular in econophysics) where, instead of the linear price response, $\Delta p _ { t } = \lambda \mathrm { s i g n } ( x _ { t } + \epsilon _ { t } ) \sqrt { | x _ { t } + \epsilon _ { t } | }$

Figure 7: Comparison of predictability performance with parametric models  
![](images/69574cc3f25e1592053b624200e95a24b8e0cc933ba6bff2188d4ecd4bd8d23e.jpg)

![](images/993f8922b14fc34d489249f43abe8c993e4ffb5b2cd5567d6d34040a80050055.jpg)  
Note: Average performance for 100 stocks return and duration predictions using RF, LASSO, Kyle, Kyle with square-root impact, Glosten-Harris, and Roll models. Each bar summarizes the mean performance over 505 days from 2019 to 2020 and over all stocks. The upper plot shows the out-of-sample $R ^ { 2 }$ , and the lower plot shows the direction accuracy, with the horizontal black line representing the baseline of a random guess.

Figure 7 shows that RF and LASSO outperform the standard microstructure-based models. This superiority is more pronounced regarding the returns $R ^ { 2 }$ than the price direction accuracy, indicating that RF and LASSO are more robust in handling noisy data compared to the four parametric models, since $R ^ { 2 }$ is a noisier measurement than direction prediction accuracy.

Besides comparing how accurate the predictions are relative to the truth for RF, LASSO, and the four parametric models, we also compute the correlations between the predictions for return direction made by each of the models to determine whether the diferent models make consistent predictions with each other. The correlation is calculated as the percentage of predictions where two methods agree on the direction, irrespectively of whether that prediction is correct or not. Figure 8 shows results for 5-second and 30-second return prediction. We find very high correlations among the predictions made by the four parametric microstructure models; correlations between RF and LASSO are also very high. However, correlations between these two groups of methods (parametric vs. machine learning) are not as high, which suggests that LASSO and RF do not simplify to the parametric structures (or to their set of features) while producing better predictions than the parametric models as seen in Figure 7. Finally, the correlations are generally lower for 30-second predictions than that for 5-second predictions, consistent with increased noise over longer time horizons.

![](images/3f1f1a526c1befefb54cded043bcca6e91ddc16208c681ac4841a7a7b9c3ac9f.jpg)

Figure 8: Correlation of direction prediction among diferent models  
![](images/0fcd5acdb2ff64c12105c7cd5ba2625073e22dc2b5ae20c679f3a327b839045b.jpg)

![](images/5bf626f11517b643d55e80ba279760d03c5b9d77f0b49a401f6e9ae6a02aa093.jpg)  
Note: Direction prediction correlation of RF, LASSO, Kyle, Kyle with square-root impact, Gllosten-Harris, and Roll model. Results are averaged over 100 stocks. The left and right heatmaps show correlation matrices for 5-second and 30-second return prediction, respectively.

## 3.6 Prediction Consistency Over Time

We have shown that with relatively little algorithm tuning, we can achieve out-of-sample $R ^ { 2 }$ for predicting returns and durations between 8 and 10% averaged across stocks and days. Is this performance achieved consistently over time, or is the average out-of-sample $R ^ { 2 }$ the skewed product of having a few days where accurate predictions compensate for poor predictions the rest of the time? To answer this question, we report the time series of day-by-day out-of-sample $R ^ { 2 }$ , along with the associated standard deviation computed across all stocks in Figure 16 in the Supplemental Material. We also report there the times series of results for the direction accuracy measure, and the out-of-sample $R ^ { 2 }$ for predicting duration. As the figure shows, the results are consistent across the sample, and the predictability measures are significantly positive. The green shaded area on the graph indicates the roughly two month period during the Spring of 2020 when the aggregate stock market dropped precipitously due to the advent of COVID-19. This period is associated with increased volatility, which results in a slight decrease in the predictability of high-frequency returns and more volatility in duration predictions.

Overall, the predictability appears to be quite stable over time. It is useful nevertheless – and this will be the object of Section 5 below – to examine in more detail what factors afect the variation in the amount of predictability, both cross-sectionally across stocks and in the time series across diferent aggregate market environments.

## 4. The Value of a Millisecond

There is a large amount of evidence that some high-frequency trading firms go to sometimes extreme lengths to obtain data as fast as possible and reduce the latency of their interactions with stock exchanges: this includes physically locating as close as possible to the exchange servers, investing in dedicated “over the horizon” or other “straight line” transmission technologies between major financial centers, etc. $^ { 1 0 } \quad \mathrm { S o } ,$ it is natural to expect that each millisecond is very valuable, or conversely every delay is costly. These technologies allow firms to quickly send and receive messages with the exchange for the purpose of placing or canceling orders and receive up-to-date data from the exchanges. Without data, no predictions are possible, and without predictions trading decisions quickly reduce to a sequence of coin tosses.

In this Section, we investigate how the amount of predictability varies with the timeliness of the data. We answer several questions related to this. First, how fast does the (fairly large) predictability identified above disappear? Second, how costly, in terms of predictability, would a delay or lag in acquiring and processing the data be? Third, is there additional predictability to be obtained from being able to look ahead, however briefly and imperfectly, at the incoming order flow?

## 4.1 The Predictability Lifespan

We have quantified above the predictability of returns, trade direction, and durations over fixed horizons. Such predictability, especially in returns, should be very short-lived in a competitive and eficient market. Especially since all the data we use are easily and publicly available, any systematic predictability that relies on it should be traded away quickly by others with similar data and models. This is indeed what we find: the predictability is short-lived.

More specifically, for predicting Return(T, ∆, M), we consider horizons ∆ =1, 3, 5, 15, 30, 60, 120, 300, 600, 1800 seconds in the calendar clock, ∆ =1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000 transactions in the transaction clock and $\Delta = 0 . 1 \mathrm { K }$ , 0.2K, 0.5K, 1K, 2K, 5K, 10K, 20K, 50K, 100K, 200K, 500K traded volume in the volume clock. The values of $\Delta$ for the duration and direction predictions are set similarly. The predictor variables in each time clock are kept the same as in Section 3, and the algorithms are tuned separately for each time horizon.

Figure 9 shows the results for each return and direction prediction. The x-axis represents diferent horizons (∆) and is plotted in log-scale. We calculate the mean over 505 dates for each stock.

The points in the plots are results averaged over all stocks, and the shaded area around each point is the 95% confidence interval of the means, measured in terms of out-of-sample $R ^ { 2 }$ or accuracy, over the stocks. They indicate the distributions of predictability of returns over the stocks for diferent time horizons. In each of the tests, we find that predictability decreases as the horizon span increases. The predictability of returns decreases monotonically as the calendar time horizon increases, but importantly remains above the baseline. In transaction and volume clocks, the predictability first increases and peaks around 10 transactions or 1K in volume. This is possibly due to a bias-variance trade-of, since the average return in the baseline becomes less noisy as more transactions are included, while additional returns included in the averaging become less and less predictable.

The results for duration predictions are shown in Figure 10. The transaction clock duration is more predictable over a longer horizon than a shorter one, consistent with the fact that transactions become more evenly distributed over longer horizons. By contrast, predictability for volume clock duration is largely stable and slowly decreases as the span increases.

## 4.2 The Impact of Delays in Acquiring or Processing Data

In all the results above, we studied the predictability of future returns with the benefit of the most up-to-date data. However, delays in data transmission and then computation are inevitable in realworld situations. What is the impact of such delays on the amount of predictability that can be

Figure 9: Predictability lifespan: Returns prediction performance as a function of the time horizon  
![](images/5a44fe8d03a0f2fa4143305d4dc2143a76b78fdb3ddb85f4a2046c71144abd2f.jpg)

![](images/29c45f3128d0b55b5c513e3567f26b82d379189e4ae32d9f1618f98c96c9b876.jpg)

![](images/8e2a89d6668d058bce64ba04cd71a970d49c915fd027d291374c59e8fbe3a6df.jpg)

![](images/4ea0bffa6a05f9ed47a788ad482b9075ea66e82e4a397caeab326a7e93580d8a.jpg)

![](images/44e14cc875247ed5271fa951b7a0c8fbee10d2181cf35f6776ae614ec2c59bcb.jpg)

![](images/9d4108d904ebc3ecc966d782a85ceb6f85991d414bf24e8a5c5b039ef1045f94.jpg)

![](images/0aa4b452df304e2329c87c9b1459bdb877532d8c02fcb62a39c066c276649dbc.jpg)

![](images/f79b6e64cc0e7beb403d0a78021865eb593d98957cd9bb5de8e6ba54af2c9d91.jpg)

![](images/d13725edbf1704deb730d71f758169597dec7bab298da1107602846e9a68a53c.jpg)  
Note: The shaded areas depict 95% confidence intervals of the mean out-of-sample $R ^ { 2 }$ or accuracy over all stocks. The upper panels are the results for return predictions in calendar, transaction, and volume clocks, respectively. The middle and lower panels are their corresponding results for direction predictions and quartile predictions, respectively. The horizontal black lines are the baselines for each problem: 0 for return $R ^ { 2 }$ using the sample mean of training set, 0.5 for direction accuracy using a random guess, and 0.25 for quartile accuracy using a random guess.

extracted from past data?<sup>11</sup>

Figure 10: Predictability lifespan: Duration prediction performance as a function of the time horizon  
![](images/2a0e21882eafb6f0507e971b2281f211af1b15c88fd0a125e282caf6c1165b67.jpg)

![](images/79db615eed9c9cc576583a786567c641634c0d4dde12415e58ead36ff4a76bc3.jpg)  
Note: Left and right panels are predictability of duration in transaction clock and volume clock respectively. The shaded area is 95% confidence intervals of the mean out-of-sample $R ^ { 2 }$ or accuracy (over all stocks). The horizontal black lines are the baselines using sample means of the training sets.

A delay can take various forms. The most up-to-date transaction messages and quote updates sent from exchange servers might take some time to reach the system of a trading algorithm. The system needs time to process the data, make decisions based on new information, and send orders based on them. Reducing computation and data processing time is expensive. These are delays in the calendar clock. In addition, a limit order might not be on the top of the limit order book when received by the exchange and will need to wait for a few transactions before being executed. Such delays can be thought of as occurring in a transaction or volume clock.

We now quantify the predictability decay when a delay in data availability and/or processing is incorporated. More specifically, letting δ denote the delay at time $T .$ , we now predict the delayed version of returns as Return $( T + \delta , \Delta , M )$ . All the predictor variables continue to be calculated at time $T .$ . Depending on the clock mode, the delay can be a few milliseconds, several transactions, or a few lots of traded volume. $\delta = 0$ corresponds to no delay and reduces to the previous results. We focus attention on the predictions of returns and trade direction at horizons $\Delta$ of 5 seconds, 10 trades, or 10 lots traded volume.

The results are shown in Figure 11. We use the same time clock for each delay and prediction problem. The delays $\delta$ are 0s, 0.1ms, 0.3ms, 1ms, 3ms, 10ms, 30ms, 0.1s, 0.3s, 1s, 3s, 5s, 10s and 30s in calendar clock, 0, 1, 3, 5, 10, 30, 150 trades in transaction clock, and 0, 1, 3, 5, 10, 30, 150 lots in volume clock. We find that the predictability decreases monotonically as the delay increases. In calendar clock, the average relative out-of-sample $R ^ { 2 }$ drops from 100% (i.e. delay-free) to 49% after a delay of 10ms (1% of a second), demonstrating the high value of the timeliness of the data on a scale of a few milliseconds. The rate of decrease then slows down, with the average relative out-ofsample $R ^ { 2 }$ dropping to 25% after a delay of 10s. Similar decreases also appear in other problems and other time clocks. The results suggest that the majority of the predictability has its source in the most recent few milliseconds of price and quote movements. Such results provide strong evidence of the value of timely data and contributes to explaining why data latency is so highly valued by high-frequency market participants.

Figure 11: The cost of data delays: Returns predictability as a function of lags in data acquisition and exploitation  
![](images/0a683661b701219bb5d9ba184da72579beb15c8c96c4f42268bb05cc449e2e04.jpg)  
Note: This figure reports the average relative predictability of returns at diferent delay levels over all stocks. The upper panel shows the decay of out-of-sample $R ^ { 2 }$ in return predictions with respect to delays in calendar, transaction, and volume clocks. The middle and lower panels demonstrate the decayed accuracy for direction predictions and quartile predictions, respectively. Each point is the averaged relative performance (w.r.t. the delay free case) of all the stocks. The shaded area indicates the 95% confidence intervals of mean daily predictability over all stocks, measured in out-of-sample $R ^ { 2 } { \mathrm { . } }$ , direction accuracy, and quartile accuracy, respectively. The horizontal black lines are the original baselines (0 for return $R ^ { 2 }$ , 0.5 for direction accuracy, and 0.25 for quartile accuracy) divided by the delay-free performance, and the horizontal red lines are delay-free performance.

## 4.3 Peeking into the Future: The Value of Signals on the Direction of the Incoming Order Flow

We just showed that delays in acquiring or processing data are very costly. How about the opposite? What if a trader were able to acquire advance information about some limited characteristics of the order flow, such as only the sign of an incoming order, most likely imperfectly, and with very minimal time to react to it.<sup>12</sup> Such knowledge about the direction of the order flow may come from diferent sources. For example, the advantage may come from the trader’s use of deeper limit order book data, from data on transactions on futures or other related securities, from the trader’s ability to interact with the market using quotes posted on other exchanges, having a direct feed to the exchanges that is faster than the publicly available one, etc.<sup>13</sup> Lewis (2015) describes many possibilities. Whether all these explanations are realistic or not, it is interesting to compute how much information of this type might help predictions.

We model this situation as follows. In addition to all the previous predictors, we now add a binary predictor that serves as a noisy estimator of the direction of the incoming order flow, averaged over a span of length $\Delta .$ Let $\mathrm { D i r } _ { t } ^ { \mathrm { L R } }$ be the binary trading direction signed using the Lee and Ready (1991) algorithm at time t and X be a Bernoulli random variable with <sup>P</sup> $\begin{array} { r } { ( X = 1 ) = p , } \end{array}$ , where $( 1 - p )$ represents the probability that the signal is actually correct. A signal of this type is an input to the optimal trading strategy by the market maker in the theoretical model of A¨ıt-Sahalia and Sa˘glam (2021). The advance signal at time $T$ is the noisy version of the average direction of future order flows, defined as

$$
\operatorname{FlowDir} (T, \Delta , M, p) = \operatorname{sign} (2 X - 1) \cdot \operatorname{sign} \Bigl (\sum_ {t \in \mathbf {D} ^ {\mathrm{txn}} \cap \operatorname{Int} ^ {\mathrm{forward}} (T, \Delta , M)} \operatorname{Dir} _ {t} ^ {\mathrm{LR}} \Bigr).\tag{4.1}
$$

This variable flips randomly the average direction of future trades with probability $( 1 - p )$ . Hence, the predictor is noiseless, predicting accurately the future trade direction, at $p = 0$ , and turns into a purely random guess of signs when $p = 0 . 5$ . Note that FlowDir is based solely on the sign of the sum of binary trade directions, without using any knowledge of price or volume information.

Figure 12: Peeking at the order flow: Predictability as a function of the accuracy of the directional signal on the incoming order flow  
![](images/19625822c3d98d14ff1b85261da2794f183385bf067856d6a7421cb303dd11bb.jpg)

![](images/5b9c8ddb8aca481ada992d000560a7591b804aca4cdd329c15c84f5be0c9456d.jpg)

![](images/98487e1a3f43c1ba86ddd9730a2d7295bb9054dae7c1b8f7eb8c2445963e4d52.jpg)  
Note: Average relative predictability at diferent noise level. The responses are 5-second returns, directions, and quartile of all stocks. The shaded area is the 95% confidence interval of mean daily predictability over all stocks. The red line depicts the (relative) average predictability in our main result, where no look-ahead signals were used. The horizontal black lines are the original baselines (0 for return $R ^ { 2 }$ , 0.5 for direction accuracy, and 0.25 for quartile accuracy) divided by the no-look-ahead signals performance.

We focus on a peeking-ahead span of $\Delta = 5$ seconds in M = calendar clock, the same as the horizon for the returns predictions. To let the algorithms make full use of such information, we allow its interactions with every other predictor already available. This doubles the total number of predictors to 270 from the previous 135. Figure 12 reports the averaged results for various values of p for predicting the stocks’ 5-second returns and their associated directions. The shaded area indicates the 95% confidence intervals of the mean over the stocks. The results show that including the sign of the average future transaction direction can boost the relative return predictability by 150%, directional accuracy by 18%, and quartile accuracy by 29% from the performance without look-ahead signals. And, as expected, the predictability increases monotonically as signals become more informative (1 − p increases). Again, we take no stand on whether traders have in reality the ability to infer the direction of the incoming order flow. What is clear, however, is that such ability is (or would be) very valuable.

## 5. Cross-Sectional and Time-Series Determinants of Predictability

In this Section, we study how the predictability we are able to achieve for returns and durations vary across stocks and days, and what variables explain this variation.

## 5.1 Nominal Share Price Level and Price Discreteness

We start by showing that price discreteness is an important factor driving the predictability of returns. The minimum price increment \$0.01 that is necessary to record a non-zero return means that a non-zero return is a larger event relative to its volatility for a stock with a lower nominal price per share, so we should expect such an event to be easier to predict. Stocks with small nominal share prices are traded with tick sizes as large as 5bps or 10bps. And since the bid-ask spread is wide and the gap between the best and second bid / ask prices is also wide, a larger proportion of orders are placed at the best bid and ask, making the estimation of buy and sell pressure from either side more accurately using the level-1 quote data, resulting in better predictions of the short-horizon return and trade direction.

Figure 13 shows a scatterplot of each stock’s average daily predictability against its average daily closing share price. The figure shows that the three stocks with the highest predictability in returns, which are Ford (F), General Electric (GE), and Kinder Morgan Inc. (KMI), are also the ones with the lowest nominal share prices. The average daily closing prices in 2019-2020 for F, GE, and KMI are \$8.10, \$9.00, and \$17.60, respectively. We record a remarkably high average daily out-of-sample $R ^ { 2 }$ of 48%, 39% and 33% when predicting their 5-second returns. By contrast, the majority of stocks in the S&P100 index have a share price around or above \$100 and an average daily out-of-sample $R ^ { 2 }$ between 5 and 15%. And there is a clear negative relationship between share prices and predictability in return and trade direction.

Figure 13: Prediction performance as a function of nominal share prices  
![](images/46951544ef26efe35422825de340461403b5a82e0be5c8b34c83c829103c80d7.jpg)

![](images/ac2d71f2190b86219b3e6f574e5fd1cb314848f8cb0fc752df8cf269df841276.jpg)

![](images/87d3c7033f463c074a5423f87bd85a08ff90f88cd2b47bfc05172393663e6769.jpg)  
Note: Each point represents a security’s average daily nominal share prices in 2019 and 2020 against its average daily performances, which are the out-of-sample $R ^ { 2 }$ for predicting 5-second returns (left panel), their associated directional accuracy (middle panel), and out-of-sample $R ^ { 2 }$ for predicting duration in 10 trades (right panel). The red lines are the OLS fits of data

On the other hand, when predicting the duration of the next 10 transactions, predictability increases for stocks with larger nominal share prices. A possible explanation is that larger nominal share price tends to correlate with more liquid trading, which makes durations easier to anticipate. Table 4 shows the result of a panel regression of predictability on log nominal share prices across all days and stocks, with days and stocks fixed efects. Each explanatory variable is normalized prior to the panel regression. The regression results confirm that nominal share price is negatively associated with return or direction predictions but positively associated with predictions of the duration of transactions, even after controlling for additional confounding factors.

## 5.2 Liquidity, Volatility, Jumps, Asset Pricing Characteristics, and Market Environment

We have also similarly examined the marginal impacts of stock trading liquidity, stock-level volatility, jumps, asset pricing characteristics, and market-wide environment on the predictability of returns, with and without adjusting symbol and date fixed efects. The results are reported in §D of the Supplemental Material. The following multivariate panel regression shows further that these efects persist even when all of them are considered.

Table 4: Regression of predictability on nominal share price level

<table><tr><td></td><td>5s return</td><td>5s direction</td><td>10 txn duration</td></tr><tr><td>Nominal Share Price (log)</td><td>-0.019***(0.001)</td><td>-0.011***(0.001)</td><td>0.006***(0.002)</td></tr><tr><td>Symbol fixed effect</td><td>YES</td><td>YES</td><td>YES</td></tr><tr><td>Date fixed effect</td><td>YES</td><td>YES</td><td>YES</td></tr><tr><td>Adjusted R2</td><td>0.698</td><td>0.686</td><td>0.144</td></tr></table>

Notation: <sup>∗</sup>p<0.1; <sup>∗∗</sup>p<0.05; <sup>∗∗∗</sup>p<0.01

## 5.3 Multivariate Stock-Level Predictability

Finally, we consider all these determinants of predictability together in a comprehensive multivariate panel regression analysis, including stock level fixed efects and date efects. The results for return and duration predictions are shown below; the results for direction prediction are not included to save space.

Table 5 shows the regression results for the stock return predictions. We include two regressions: in the first, fixed efects control for all dates, while in the second fixed efects control only for the dates with material economic data releases. These include 14 FOMC dates and 24 employment data dates from 2019 to 2020, resulting in 38 variables. See §D.4 for additional details.

The impact of most of the explanatory variables is similar to the previously identified univariate efect. Also, the results are mostly consistent across diferent lengths of windows and diferent time clocks. As a quick summary, the returns and trade directions of less liquid and less volatile securities with smaller nominal share prices and less correlated with the market are on average easier to predict.

Table 12 in the Supplemental Material gives the regression results for the duration predictions with two regressions corresponding to diferent sets of fixed efects for each problem, as in the previous table. Both liquidity measures and volatility measures have clear positive relationships with duration predictability, while share prices, market capitalization, beta, $R ^ { 2 }$ relative to the market are negatively related to duration predictability.

m<sup>inants</sup> <sup>of</sup> <sup>predictability</sup> <sup>for</sup> <sup>returns:</sup> <sup>Panel</sup> <sup>regression</sup> <sup>of</sup> <sup>ou</sup>

<table><tr><td rowspan="2"></td><td colspan="12">Dependent variable: out-of-sample  $R^2$ s for Return Predictions</td></tr><tr><td>5 sec</td><td>5 sec</td><td>30 sec</td><td>30 sec</td><td>10 txn</td><td>10 txn</td><td>200 txn</td><td>200 txn</td><td>1K vol</td><td>1K vol</td><td>20K vol</td><td>20K vol</td></tr><tr><td>Nominal Share Price (log)</td><td>0.004*(0.002)</td><td>-0.002(0.002)</td><td>0.004***(0.001)</td><td>0.002(0.001)</td><td>-0.012***(0.003)</td><td>-0.020***(0.003)</td><td>-0.020***(0.002)</td><td>-0.021***(0.002)</td><td>-0.029***(0.003)</td><td>-0.034***(0.003)</td><td>-0.038***(0.002)</td><td>-0.040***(0.002)</td></tr><tr><td>Daily Return (log)</td><td>0.00002(0.0003)</td><td>0.001*(0.0003)</td><td>0.001***(0.0002)</td><td>0.001***(0.0002)</td><td>-0.002***(0.0003)</td><td>-0.001***(0.0004)</td><td>-0.0003(0.0002)</td><td>-0.0002(0.0002)</td><td>-0.002***(0.0003)</td><td>-0.001***(0.0003)</td><td>-0.0005(0.0003)</td><td>-0.0005(0.0003)</td></tr><tr><td>Traded Volume (log)</td><td>-0.003***(0.001)</td><td>-0.004***(0.001)</td><td>-0.007***(0.0005)</td><td>-0.005***(0.0004)</td><td>0.027***(0.001)</td><td>0.024***(0.001)</td><td>0.018***(0.0005)</td><td>0.017***(0.0004)</td><td>0.029***(0.001)</td><td>0.027***(0.001)</td><td>0.018***(0.001)</td><td>0.018***(0.001)</td></tr><tr><td>Total Market Cap (log)</td><td>-0.031***(0.002)</td><td>-0.009***(0.002)</td><td>-0.007***(0.001)</td><td>-0.001(0.001)</td><td>-0.021***(0.002)</td><td>0.006**(0.002)</td><td>-0.003***(0.001)</td><td>0.004***(0.001)</td><td>-0.017***(0.002)</td><td>0.003(0.002)</td><td>-0.005**(0.002)</td><td>-0.0004(0.002)</td></tr><tr><td>Proportional Spread</td><td>0.009***(0.0003)</td><td>0.005***(0.0003)</td><td>0.006***(0.0002)</td><td>0.004***(0.0002)</td><td>0.009***(0.0004)</td><td>0.005***(0.0003)</td><td>0.005***(0.0002)</td><td>0.005***(0.0002)</td><td>0.008***(0.0003)</td><td>0.006***(0.0003)</td><td>0.006***(0.0003)</td><td>0.005***(0.0003)</td></tr><tr><td>Volatility</td><td>-0.009***(0.001)</td><td>-0.014***(0.001)</td><td>-0.004***(0.001)</td><td>-0.002***(0.0004)</td><td>-0.008***(0.001)</td><td>-0.017***(0.001)</td><td>-0.002***(0.001)</td><td>-0.007***(0.0004)</td><td>-0.008***(0.001)</td><td>-0.011***(0.001)</td><td>-0.007***(0.001)</td><td>-0.008***(0.001)</td></tr><tr><td>Beta with SPY</td><td>0.014***(0.001)</td><td>0.001*(0.0004)</td><td>0.007***(0.0005)</td><td>0.0001(0.0003)</td><td>0.014***(0.001)</td><td>0.002***(0.001)</td><td>0.005***(0.0005)</td><td>0.001***(0.0003)</td><td>0.014***(0.001)</td><td>0.002***(0.0005)</td><td>0.003***(0.001)</td><td>-0.001(0.0004)</td></tr><tr><td>R^2 with SPY</td><td>-0.043***(0.001)</td><td>-0.020***(0.0004)</td><td>-0.026***(0.001)</td><td>-0.014***(0.0003)</td><td>-0.052***(0.001)</td><td>-0.030***(0.001)</td><td>-0.018***(0.001)</td><td>-0.009***(0.0003)</td><td>-0.051***(0.001)</td><td>-0.029***(0.0005)</td><td>-0.015***(0.001)</td><td>-0.007***(0.0004)</td></tr><tr><td>Idiosyncratic Volatility</td><td>-0.034***(0.001)</td><td>-0.011***(0.001)</td><td>-0.021***(0.001)</td><td>-0.015***(0.0005)</td><td>-0.030***(0.001)</td><td>-0.005***(0.001)</td><td>-0.015***(0.001)</td><td>-0.002***(0.0005)</td><td>-0.027***(0.001)</td><td>-0.008***(0.001)</td><td>-0.005***(0.001)</td><td>-0.0004(0.001)</td></tr><tr><td>Jump 3-4%</td><td>-0.012***(0.001)</td><td>-0.016***(0.001)</td><td>-0.003***(0.001)</td><td>-0.004***(0.001)</td><td>-0.015***(0.001)</td><td>-0.018***(0.001)</td><td>-0.002***(0.001)</td><td>-0.002***(0.001)</td><td>-0.015***(0.001)</td><td>-0.015***(0.001)</td><td>-0.002*(0.001)</td><td>-0.003***(0.001)</td></tr><tr><td>Jump 4-5%</td><td>-0.012***(0.001)</td><td>-0.021***(0.001)</td><td>-0.002**(0.001)</td><td>-0.004***(0.001)</td><td>-0.019***(0.002)</td><td>-0.028***(0.002)</td><td>-0.002*(0.001)</td><td>-0.003***(0.001)</td><td>-0.018***(0.002)</td><td>-0.024***(0.002)</td><td>-0.003*(0.002)</td><td>-0.005***(0.001)</td></tr><tr><td>Jump 5%+</td><td>-0.021***(0.001)</td><td>-0.031***(0.001)</td><td>-0.007***(0.001)</td><td>-0.010***(0.001)</td><td>-0.030***(0.002)</td><td>-0.043***(0.002)</td><td>-0.004***(0.001)</td><td>-0.007***(0.001)</td><td>-0.028***(0.002)</td><td>-0.036***(0.001)</td><td>-0.005***(0.001)</td><td>-0.008***(0.001)</td></tr><tr><td>S&amp;P500 Return</td><td></td><td>0.001*(0.0003)</td><td></td><td>-0.0002(0.0002)</td><td></td><td>0.003***(0.0004)</td><td></td><td>0.0005**(0.0002)</td><td></td><td>0.002***(0.0003)</td><td></td><td>-0.0001(0.0003)</td></tr><tr><td>VIX</td><td></td><td>0.022***(0.001)</td><td></td><td>0.010***(0.0003)</td><td></td><td>0.023***(0.001)</td><td></td><td>0.009***(0.0003)</td><td></td><td>0.017***(0.001)</td><td></td><td>0.004***(0.001)</td></tr><tr><td>Symbol fixed effect</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td></tr><tr><td>Date fixed effect</td><td>YES</td><td>NO</td><td>YES</td><td>NO</td><td>YES</td><td>NO</td><td>YES</td><td>NO</td><td>YES</td><td>NO</td><td>YES</td><td>NO</td></tr><tr><td>Data Date fixed effect</td><td>NO</td><td>YES</td><td>NO</td><td>YES</td><td>NO</td><td>YES</td><td>NO</td><td>YES</td><td>NO</td><td>YES</td><td>NO</td><td>YES</td></tr><tr><td>Adjusted  $R^2$ </td><td>0.728</td><td>0.668</td><td>0.740</td><td>0.716</td><td>0.755</td><td>0.701</td><td>0.768</td><td>0.756</td><td>0.690</td><td>0.646</td><td>0.748</td><td>0.744</td></tr></table>

## 6. Robustness Checks

We conclude the analysis with a series of robustness checks to understand how the results might change with respect to some modeling details. First, we examine the impact of the choice of algorithm by including a wider range of prediction methods, such as ridge regression. Second, we investigate the predictability using additional data from correlated stocks. Third, we check whether there are any intraday seasonality patterns in the amount of predictability at diferent times of the day. In addition, we compute how much additional predictability can be coaxed out of the random forests by increasing the number of trees and quantify the loss of predictability that would result from limiting the predictors to be statistics derived from the transaction data alone, or the quote data alone, as opposed to being able to use and combine both. In addition, we also depict incremental predictability using additional data from correlated stocks. These results are reported in $\ S \mathrm { F }$ of the Supplemental Material.

## 6.1 Comparison and Consistency of Results Across Prediction Methods

We start by checking the robustness of the prediction results by conducting a horse race against a wider range of prediction methods. For predicting returns and their signs, we include five additional prediction methods: (i) linear methods in the form of ordinary least square linear regression (OLS) and ridge regression; (ii) FarmPredict, which first decomposes observed data into factors and idiosyncratic components and uses them as new predictors in a statistical machine learning method (Fan et al., 2020a); in addition to using the learned factors, we apply LASSO on the idiosyncratic components to potentially enhance the predictive power; (iii) nonlinear methods: gradient boosted trees.

Results are shown in Figure 14. In a nutshell, all methods have very similar performance, except OLS. The improvements from a nonlinear method like random forest or gradient boosting trees are limited for most cases when compared with LASSO. The poor performance of OLS for several tests is due to heavy overfitting, with predictions that are sensitive to noise and result in very far-fetched predictions on several days. We would like to remark that, unlike other models where tuning and fitting are easy, and foolproof, fitting a working feed-forward neural network (FNN) in our return prediction problem is much more challenging. The method is very sensitive to initial states and tuning variables and outputs far-fetched predictions occasionally from time to time. For these reasons, we did not present the result here.

Figure 14: Comparison of predictability performance using diferent machine learning methods  
![](images/560b39645c0ed3da7fbf413e91a2d3b217bcedab4cbeda3e7a88cee2f04b8406.jpg)  
Note: Average performance for INTC’s return and duration predictions using OLS, LASSO, ridge regression, FarmPredict with Lasso, Random Forest (RF), and Gradient Boosted Trees (GBT). Each bar summarizes the mean performance over 505 days from 2019 to 2020.

## 6.2 Intraday Seasonality: Predictability Across Diferent Trading Hours

Finally, we examine whether intraday seasonality plays a role in the predictability over a trading session. It is well documented that volatility is typically higher at the beginning and the end of trading days and lower in the middle, and spreads tend to decrease throughout the day. In light of these stylized facts, we divide the trading day into three disjoint time intervals, a 9:30-10:00 opening session, 10:00-15:30 midday session and 15:30-16:00 closing session. For each fraction of the day, we tune, fit, and test models using data only from the same time intervals in the past days using the same procedure as before.

The results are shown in Figure 15. Returns tend to be easier to predict in the middle of the day rather than at the beginning of the day, possibly owing to the large moves typically concentrated during the overnight hours, while durations are similarly predictable. Interestingly, predictability at the end of the day is higher for all of the prediction problems we considered, despite the higher volatility, suggesting that the trading patterns at the end of the day are more consistent across days.

## 7. Conclusion

We studied three basic prediction problems over ultra short high-frequency horizons, namely predictions of returns, transaction direction and durations. We examined them in three time clocks using primarily LASSO and random forests, and relying on the complete transaction and quote update data of the S&P 100 stocks over two full years from 2019 to 2020.

Figure 15: Intraday seasonality: Predictability across diferent phases of the trading day  
![](images/327ca78b3999f3f5bf77bac45821e47df53d973be1b6d231d3e086b9fbdc26ad.jpg)  
Note: Each bar is the average performance (out-of-sample $R ^ { 2 } )$ of random forest in the problem. Each model is fitted and tested with data from the same specified range of trading hours.

We quantified the predictability and found that fairly large amounts of predictability exist universally in every stock and do so consistently over time, a situation very diferent from long horizon, low frequency predictability. We also isolated the variables that were most responsible for driving this predictability: they include the transaction imbalance, past returns and limit order book imbalance for return and direction predictions, and volume statistics for duration predictions.

We then investigated how the predictability depends on stock characteristics and market environments: for return or direction predictions, securities with less liquid, less volatile, smaller nominal share prices, and those weakly correlated with the market are more predictable. By comparison, predictability for durations is higher under liquid and volatile conditions.

We examined next the value of the timeliness of the data and found that predictability of returns and trading directions vanishes quickly as soon as the time span is larger than a few minutes, a few thousand trades, or a few thousand lots traded. In addition, the majority of the predictability lies in the most up-to-date few milliseconds, or a few trades, and decreases quickly when a small delay is introduced. We also simulated the potential ability of high-frequency traders to make short-term imperfect predictions on the sign of the incoming order flow. Such ability turned out to be very useful in prediction, almost doubling the out-of-sample $R ^ { 2 }$ in 5-second return predictions.

Finally, we found that other statistical machine learning algorithms, such as ridge regression or gradient boosting trees produce substantially the same results. OLS however does not identify any predictability, highlighting the value of the penalization for overfitting that characterizes most machine learning algorithms. Fine-tuning a given algorithm, such as increasing the number of trees in random forests, reaches its limits beyond a certain point. Transaction data are more valuable than quote data in terms of predicting returns. Incorporating data from additional correlated stocks does not have a large impact on the predictability of one stock: its own transactions and quotes are by far the most informative. Finally, there exists some intraday variability in the amount of predictability, with the end of the day subperiod being the most predictable.

Acknowledgment. The authors gratefully acknowledge Dr. Yifeng Zhou’s significant help and contributions at an earlier stage of this project. Fan’s research was supported by NSF grants DMS-2210833 and DMS-2412029.

## References

A¨ıt-Sahalia, Y., Brunetti, C., 2020. High frequency traders and the price process. Journal of Econometrics 217, 20–45.

A¨ıt-Sahalia, Y., Sa˘glam, M., 2021. High frequency market making: The role of speed. Tech. rep., Princeton University.

Alvim, L. G., dos Santos, C. N., Milidiu, R. L., 2010. Daily volume forecasting using high frequency predictors. Tech. rep., Pontificia Universidade Catolica do Rio de Janeiro.

Baron, M., Brogaard, J. A., Hagstr¨omer, B., Kirilenko, A., 2019. Risk and return in high frequency trading. Journal of Financial and Quantitative Analysis 54, 993–1024.

Bickel, P. J., Ritov, Y., Tsybakov, A. B., 2009. Simultaneous analysis of lasso and dantzig selector. The Annals of Statistics 37 (4), 1705–1732.

Chinco, A., Clark-Joseph, A. D., Ye, M., 2019. Sparse signals in the cross-section of returns. The Journal of Finance 74, 449–492.

Cont, R., Kukanov, A., Stoikov, S., 2014. The price impact of order book events. Journal of Financial Econometrics 12, 47–88.

Cont, R., Stoikov, S., Talreja, R., 2010. A stochastic model for order book dynamics. Operations Research 58, 549–563.

Dixon, M. F., 2018. A high-frequency trade execution model for supervised learning. High Frequency 1, 35–52.

Easley, D., de Prado, M. L., O’Hara, M., Zhang, Z., 2021. Microstructure in the machine age. Review of Financial Studies 34, 3316–3363.

Fama, E. F., 1970. Eficient capital markets: A review of theory and empirical work. The Journal of Finance 25, 34–105.

Fan, J., Ke, Y., Wang, K., 2020a. Factor-adjusted regularized model selection. Journal of Econometrics 216, 71–85.

Fan, J., Li, R., Zhang, C.-H., Zou, H., 2020b. Statistical Foundations of Data Science. Chapman and Hall/CRC.

Fan, J., Wang, W., Zhu, Z., 2021. A shrinkage principle for heavy-tailed data: High-dimensional robust low-rank matrix recovery. Annals of Statistics 49, 1239–1266.

Glosten, L. R., Harris, L. E., 1988. Estimating the components of the bid/ask spread. Journal of Financial Economics 21, 123–142.

Hagstr¨omer, B., 2021. Bias in the efective bid-ask spread. Journal of Financial Economics 142, 314–337.

Hastie, T., Tibshirani, R., Friedman, J., 2009. The Elements of Statistical Learning: Data Mining, Inference, and Prediction, 2nd Edition. Springer Series in Statistics. Springer-Verlag, New York.

Huang, R. D., Stoll, H. R., 1994. Market microstructure and stock return predictions. Review of Financial Studies 7, 179–213.

Kercheval, A. N., Zhang, Y., 2015. Modelling high-frequency limit order book dynamics with support vector machines. Quantitative Finance 15, 1315–1329.

Knoll, J., St¨ubinger, J., Grottke, M., 2019. Exploiting social media with higher-order factorization machines: Statistical arbitrage on high-frequency data of the S&P 500. Quantitative Finance 19, 571–585.

Lee, C. M., Ready, M. J., 1991. Inferring trade direction from intraday data. The Journal of Finance 46, 733–746.

Lewis, M., 2015. Flash Boys: A Wall Street Revolt. W. W. Norton & Company.

Lo, A. W., MacKinlay, A. C., 2002. A Non-Random Walk Down Wall Street. Princeton University Press.

Malkiel, B. G., 1973. A Random Walk Down Wall Street. W. W. Norton & Company.

Moallemi, C., Sa˘glam, M., 2013. The cost of latency in high-frequency trading. Operations Research 61, 1070–1086.

Murphy, K. P., 2014. Machine Learning: A Probabilistic Perspective. MIT Press.

Ntakaris, A., Magris, M., Kanniainen, J., Gabbouj, M., Iosifidis, A., 2018. Benchmark dataset for mid-price forecasting of limit order book data with machine learning methods. Journal of Forecasting 37, 852–866.

O’Hara, M., Yao, C., Ye, M., 2014. What’s not there: Odd lots and market data. The Journal of Finance 69, 2199–2236.

Panayi, E., Peters, G. W., Danielsson, J., Zigrand, J.-P., 2018. Designating market maker behaviour in limit order book markets. Econometrics and Statistics 5, 20–44.

Roll, R., 1984. A simple model of the implicit bid-ask spread in an eficient market. The Journal of Finance 39, 1127–1139.

Ro¸su, I., 2009. A dynamic model of the limit oder book. Review of Financial Studies 22, 4601–4641.

Sirignano, J. A., 2019. Deep learning for limit order books. Quantitative Finance 19, 549–570.

Timmermann, A., 2018. Forecasting methods in finance. Annual Review of Financial Economics 10, 449–479.

Tsantekidis, A., Passalis, N., Tefas, A., Kanniainen, J., Gabbouj, M., Iosifidis, A., 2017. Forecasting stock prices from the limit order book using convolutional neural networks. In: 2017 IEEE 19th Conference on Business Informatics (CBI). Vol. 1. IEEE, pp. 7–12.

Zhang, L., Mykland, P. A., A¨ıt-Sahalia, Y., 2005. A tale of two time scales: Determining integrated volatility with noisy high-frequency data. Journal of the American Statistical Association 100, 1394–1411.

Zhao, P., Yu, B., 2006. On model selection consistency of lasso. The Journal of Machine Learning Research 7, 2541–2563.

Zheng, B., Moulines, E., Abergel, F., 2013. Price jump prediction in a limit order book. Journal of Mathematical Finance 3, 242–255.

## A. Summary statistics of TAQ data

Table 6: Basic description of the size of TAQ data used

<table><tr><td>Securities included</td><td>all companies in S&amp;P100 on 2020.12.31</td></tr><tr><td>Number of different security symbols</td><td>101</td></tr><tr><td>Date range of data</td><td>2019.1.1 to 2020.12.31</td></tr><tr><td>Number of trading days included</td><td>505</td></tr><tr><td>Number of symbols available on all days</td><td>96</td></tr><tr><td>Number of available (symbol, date) pairs</td><td>50,273</td></tr><tr><td>Total disk size</td><td>2.3 Terabytes</td></tr></table>

Summary statistics of our data and response variables are reported in Table 7. The table is divided into two parts. The upper panel contains a summary at the daily level (505 trading days) aggregated across all 101 stocks, with each stock contributing one observation each day it is available. The total market capitalization, nominal stock price, and daily returns are all based on daily closing prices. The daily beta of the stock is estimated by regressing the stock’s 15 second returns on the 15 second returns of SPY (an ETF tracking the S&P500 index) and $R ^ { 2 }$ is the coeficient of determination in this regression.<sup>14</sup> The turnover rate is the ratio of traded volume multiplied by the daily closing price over market capitalization. The lower panel of the table contains summary statistics for the response variables computed over all nanosecond-level timestamps and all stocks. We downsample the data so that each stock and day are approximately equally represented. For each pair of stock and date, 1, 000 response variables are sampled, and the data from all 50, 273 such pairs are merged to compute the summary statistics. We can observe that the high-frequency returns are largely symmetrically distributed but with very heavy tails (large kurtosis). The duration variables are both skewed and heavy-tailed.

## B. Predictor Variables

To examine how well the response variables just described can be predicted, we consider a large number of predictor (or independent) variables representing a broad range of features indicative of the short-term trading environment for a particular stock. For now, we use only a given stock’s variables to make predictions about that stock. In practice, it is plausible that predictions could be improved even further by exploiting sectors or industry-level correlation patterns to make predictions about a stock from preceding observations about another stock with correlated patterns, in what is often called “ statistical arbitrage.” We quantify the incremental predictability for a given stock that can be achieved by using correlated stocks’ data in Section F of this supplemental material.

Just like the response variables, the predictor variables we use can all be constructed using exclusively the complete record of time-stamped transactions and quote updates. They take the form of (possibly nonlinear) transformations of the transactions and quote data recorded before the time T at which the prediction takes place.

For each variable, an interesting question lies in determining the length of the lookback window that precedes T . We use a large set of disjoint such windows covering the most recent records (on a scale of seconds) all the way to far back in time (on a scale of hours) so as not to prejudge the results in favor of a short lookback window. Furthermore, there is no reason to assume that the length of the most informative lookback window should be the same for each predictor variable. And, in principle, predictors derived under one time clock can be useful for predicting a response variable measured under a diferent time clock. So there is a wide range of possible combinations, making machine learning algorithms (as opposed to traditional forecasting methods) well suited to the problem.

Similar to the forward-looking intervals (2.2) for the response variables, we construct lookback intervals to build predictor variables. For calendar time at the timestamp T , lookback spans $( \Delta _ { 1 } , \Delta _ { 2 } ) , \Delta _ { 1 } \leq \Delta _ { 2 }$ and time clock M, we define:

$$
\operatorname{Int} ^ {\text { back }} (T, \Delta_ {1}, \Delta_ {2}, M) = \left\{ \begin{array}{l l} \operatorname{Int} (T - \Delta_ {2}, T - \Delta_ {1}) & \text { if } M = \text { calendar } \\ \left\{t: t \leq T, \Delta_ {1} \leq \left(\sum_ {s \in \operatorname{Int} (t, T)} \mathbb {1} _ {\{V _ {s} > 0 \}}\right) <   \Delta_ {2} \right\} & \text { if } M = \text { transaction } \\ \left\{t: t \leq T, \Delta_ {1} \leq \left(\sum_ {s \in \operatorname{Int} (t, T)} V _ {s}\right) <   \Delta_ {2} \right\} & \text { if } M = \text { volume } \end{array} \right..\tag{B.1}
$$

For each timestamp T and clock mode M, we set the lookback spans $( \Delta _ { 1 } , \Delta _ { 2 } )$ to create different features at T with multiple disjoint intervals. For the calendar clock $( M \ = \ \mathrm { c a l e n d a r } )$ nine look-back windows are used: $( \Delta _ { 1 } , \Delta _ { 2 } ) \ \in \ \{ ( 0 , . 1 ) , ( . 1 , . 2 ) , ( . 2 , . 4 ) , . . . , ( 1 2 . 8 , 2 5 . 6 ) \}$ with number of seconds as the unit; for the transaction clock (M =transaction), the 9 spans are $( \Delta _ { 1 } , \Delta _ { 2 } ) \in$ $\{ ( 0 , 1 ) , ( 1 , 2 ) , ( 2 , 4 ) , \ldots , ( 1 2 8 , 2 5 6 ) \}$ using number of transactions as a unit. Similarly, the spans for the volume clock (M =volume) are set as $( \Delta _ { 1 } , \Delta _ { 2 } ) \in \{ ( 0 , 1 0 0 ) , ( 1 0 0 , 2 0 0 ) , ( 2 0 0 , 4 0 0 ) , \ldots , ( 1 2 8 0 0 , 2 5 6 0 0 ) \}$ , with units in number of shares, where 1 lot consists of 100 shares.

We consider 15 main predictors, each being implemented over the 9 time spans so that $9 \times 1 5 = 1 3 5$ variables are used for each of the 3 time clocks. Furthermore, the predictors are allowed to interact in multiple nonlinear fashion, selected by the machine as part of the prediction algorithm. The 15 predictors can be grouped into 3 categories, which we now describe.

Volume and duration. The first group of predictors is related to a stock’s trading intensity. One might expect for example that a higher than normal occurrence of block trades or very frequent transactions may persist over short horizons and therefore have predictive power.

1. Breadth is the number of transactions in the interval:

$$
\mathrm{Breadth} (T, \Delta_ {1}, \Delta_ {2}, M) = | \mathbf {D} ^ {\mathrm{txn}} \cap \mathrm{Int} ^ {\mathrm{back}} (T, \Delta_ {1}, \Delta_ {2}, M) |.\tag{B.2}
$$

2. Immediacy is the average time between successive transactions in the interval:

$$
\text { Immediacy } (T, \Delta_ {1}, \Delta_ {2}, M) = \frac {\Delta_ {2} - \Delta_ {1}}{\text { Breadth } (T , \Delta_ {1} , \Delta_ {2} , M)}\tag{B.3}
$$

3. VolumeAll is the total number of shares transacted in the interval:<sup>15</sup>

$$
\operatorname{VolumeAll} (T, \Delta_ {1}, \Delta_ {2}, M) = \sum_ {t \in \operatorname{Int} ^ {\text { back }} (T, \Delta_ {1}, \Delta_ {2}, M)} V _ {t}.\tag{B.4}
$$

4. VolumeAvg is the average number of shares transacted for each transaction in the interval:

$$
\operatorname{VolumeAvg} (T, \Delta_ {1}, \Delta_ {2}, M) = \frac {\operatorname{VolumeAll} (T , \Delta_ {1} , \Delta_ {2} , M)}{\operatorname{Breadth} (T , \Delta_ {1} , \Delta_ {2} , M)}.\tag{B.5}
$$

5. VolumeMax is the maximum number of shares transacted in one transaction in the interval:

$$
\operatorname{VolumeMax} (T, \Delta_ {1}, \Delta_ {2}, M) = \max \left\{V _ {t}: t \in \operatorname{Int} ^ {\text { back }} (T, \Delta_ {1}, \Delta_ {2}, M) \right\}.\tag{B.6}
$$

Return and imbalance. The second group of predictors is related to the stock’s recent trading asymmetry. For example, if a majority of trades are buy trades that hit the limit sell, or the bid is dominating the ask in the level 1 quotes, then we might expect to see an upward pressure on the price. It is natural to expect that an element in predicting future returns and durations will be characteristics of the current limit order book (LOB), including any imbalances; such imbalances are known to be indicative of future price movements (see, e.g., Cont et al. (2014) and Kercheval and Zhang (2015)). We define the following variables:

1. Lambda is the price change in the interval relative to total volume:

Let $\mathbf { I } = \mathbf { D } ^ { \mathrm { t x n } } \ \cap \mathrm { I n t } ^ { \mathrm { b a c k } } ( T , \Delta _ { 1 } , \Delta _ { 2 } , \mathrm { M } )$ , then

$$
\operatorname{Lambda} (T, \Delta_ {1}, \Delta_ {2}, M) = \frac {P _ {\max (\mathbf {I})} - P _ {\min (\mathbf {I})}}{\operatorname{VolumeAll} (T , \Delta_ {1} , \Delta_ {2} , M)}.\tag{B.7}
$$

2. LobImbalance is the average imbalance in the depth of the limit order book over the lookback interval:

$$
\mathrm{LobImbalance} (T, \Delta_ {1}, \Delta_ {2}, M) = \mathrm{Average} \left[ \frac {S _ {t} ^ {a} - S _ {t} ^ {b}}{S _ {t} ^ {a} + S _ {t} ^ {b}}: t \in \mathrm{Int} ^ {\mathrm{back}} (T, \Delta_ {1}, \Delta_ {2}, \mathrm{M}) \right].\tag{B.8}
$$

3. TxnImbalance measures the asymmetry of buy and sell volumes in recent transactions. Denote by $\mathrm { D i r } _ { t } ^ { \mathrm { L R } }$ the binary transaction direction at time t signed using the algorithm of Lee and Ready (1991). Then transaction imbalance is calculated as

$$
\mathrm{TxnImbalance} (T, \Delta_ {1}, \Delta_ {2}, M) = \frac {\sum_ {t \in \mathbf {D} ^ {\mathrm{txn}} \cap \mathrm{Int} ^ {\mathrm{back}} (T , \Delta_ {1} , \Delta_ {2} , M)} \left(V _ {t} \cdot \mathrm{Dir} _ {t} ^ {\mathrm{LR}}\right)}{\mathrm{VolumeAll} (T , \Delta_ {1} , \Delta_ {2} , M)}.\tag{B.9}
$$

4. PastReturn is the past return in the interval. It is defined similarly to the transaction return response, except over a lookback window. Let $\mathbf { I } = \mathbf { D } ^ { \mathrm { t x n } } \ \cap \mathrm { I n t } ^ { \mathrm { b a c k } } ( T , \Delta _ { 1 } , \Delta _ { 2 } , \mathrm { M } )$ , then<sup>16</sup>

$$
\mathrm{PastReturn} (T, \Delta_ {1}, \Delta_ {2}, M) = 1 - \mathrm{Average} \left[ P _ {t} ^ {\mathrm{txn}}: t \in \mathbf {I} \right] / P _ {\max (\mathbf {I})}.\tag{B.10}
$$

Speed and cost. The final set of predictors we employ measure the speed and cost inherent in the stock’s trading.

1. Turnover is the speed of transactions in relation to the stock’s total number of shares outstanding (denoted as S).

$$
\operatorname{Turnover} (T, \Delta_ {1}, \Delta_ {2}, M) = \frac {\operatorname{VolumeAll} (T , \Delta_ {1} , \Delta_ {2} , M)}{S}.\tag{B.11}
$$

2. AutoCov is the autocovariance of transaction returns in the interval. For any $t \in \mathbf { D } ^ { \mathrm { t x n } }$ , denote by $L t = \operatorname { a r g m a x } _ { s } \left\{ s : s < t , s \in \mathbf { D } ^ { \operatorname { t x n } } \right\}$ the timestamp of the transaction right before time t. Then the autocovariance is

$$
\begin{array}{l l} \text {AutoCov} (T, \Delta_ {1}, \Delta_ {2}, M) = & \text {Average} \left[ \log \left(\frac {P _ {t} ^ {\text {txn}}}{P _ {L t} ^ {\text {txn}}}\right) \log \left(\frac {P _ {L t} ^ {\text {txn}}}{P _ {L (L t)} ^ {\text {txn}}}\right): \right. \\ & \left. t \in \mathbf {D} ^ {\text {txn}} \cap \text {Int} ^ {\text {back}} (T, \Delta_ {1}, \Delta_ {2}, M) \right]. \end{array}\tag{B.12}
$$

3. QuotedSpread is the average proportional nominal spread in the quotes over the lookback interval:

$$
\operatorname{QuotedSpread} (T, \Delta_ {1}, \Delta_ {2}, M) = \operatorname{Average} \left[ \frac {P _ {t} ^ {a} - P _ {t} ^ {b}}{P _ {t}}: t \in \operatorname{Int} ^ {\text { back }} (T, \Delta_ {1}, \Delta_ {2}, M) \right].\tag{B.13}
$$

4. EfectiveSpread is the dollar-weighted percent efective spread over the interval:

EfectiveSpread

$$
\left. ^ {\prime}, \Delta_ {1}, \Delta_ {2}, M\right) = \frac {\sum_ {t \in \mathbf {D} ^ {\mathrm{txn}} \cap \operatorname{Int} ^ {\mathrm{back}} (T , \Delta_ {1} , \Delta_ {2} , \mathrm{M})} \left[ \log \left(\frac {P _ {t} ^ {\mathrm{txn}}}{P _ {t}}\right) \cdot \operatorname{Dir} _ {t} ^ {\mathrm{LR}} \cdot V _ {t} \cdot P _ {t} ^ {\mathrm{txn}} \right]}{\sum_ {t \in \mathbf {D} ^ {\mathrm{txn}} \cap \operatorname{Int} ^ {\mathrm{back}} (T , \Delta_ {1} , \Delta_ {2} , \mathrm{M})} \left(V _ {t} \cdot P _ {t} ^ {\mathrm{txn}}\right)}.\tag{B.14}
$$

5. RealizedVolatility is the local variance of transaction returns over the lookback interval:

$$
\mathrm{RV} (T, \Delta_ {1}, \Delta_ {2}, M) = \mathrm{Average} \left[ \left(\log P _ {t} ^ {\mathrm{txn}} - \log P _ {L t} ^ {\mathrm{txn}}\right) ^ {2}: t \in \mathrm{Int} ^ {\mathrm{back}} (T, \Delta_ {1}, \Delta_ {2}, \mathrm{M}) \right].\tag{B.15}
$$

6. TSRV (Two Scale Realized Volatility, see Zhang et al. (2005)) is a noise-robust volatility estimator obtained by computing a set of $\mathrm { R V } ( T , \Delta _ { 1 } , \Delta _ { 2 } , M )$ at a lower sampling frequency (replacing Lt with a time further lagged), with diferent starting points to cover the lookback interval, and averaging RV over that set.

## C. Results on Average daily performance

Here we summarize the distribution of out-of-sample $R ^ { 2 }$ for each given day across 100 stocks. The results are in Figure 16.

## D. Additional Results on Determinants of Predictability

Here, we furnish additional results on the marginal contributions of trading liquidity, volatility, jumps, and asset pricing characteristics of individual stocks to the predictability of their returns, with and without adjusting date and symbol efects, depicted respectively in the tables and figures below. In addition, we also examine the impact of the market environment on the predictability of stock returns.

## D.1 Stock Trading Liquidity

We then examine the efect of the liquidity of individual stocks on their individual predictability. We use two separate measures of liquidity: total traded dollar volume and percentage spread. For a given stock and date, the total traded volume is calculated as the total number of shares traded in a day times the closing price. The percentage spread is calculated as the average of the stock’s bid-ask spread divided by its midprice, sampled every 15 seconds throughout the day. It is natural to expect that the markets for more liquid stocks are more eficient in terms of price discovery, with past signals being incorporated faster into prices, which should make the prediction of returns (i.e., future prices) more dificult.

Figure 17 shows the stock-by-stock results, with daily predictability averaged over the full sample. A clear pattern emerges: better liquidity, in the form of higher volume or lower spreads, leads to weaker predictability in return and trade direction, whereas durations are easier to predict when liquidity is higher. The univariate relationships are confirmed by multivariate panel regressions in Table 8, with fixed efects and controls. For the same reason, market capitalization of the stock is also found to be negatively correlated with predictability in most cases, likely due to its positive relationship with liquidity.

One reason durations are more dificult to predict for less liquid stocks is that there are larger outliers in durations, in the forms of (relatively) long gaps in trading, that negatively impact the overall predictability. On the other hand, for stocks where there is relatively little liquidity available at the best bid and ask, traders are likely to break up a large order over many exchanges simultaneously to capture all available liquidity that is spread out over multiple exchanges, leading to severa transactions. By contrast, the same trader behavior in stocks with plenty of liquidity available (perhaps due to a low nominal share price and a binding minimum tick) may lead to a single trade on one exchange. But we note that even for the least predictable stock, durations remain quite predictable with out-of-sample $R ^ { 2 }$ above 7% despite a limited training set we are providing to the algorithms and no attempt at tuning the algorithms for this specific problem.

## D.2 Stock-Level Volatility and Jumps

It is natural to expect that cross-sectional diferences in volatility and jump intensity might afect predictability. If the trading conditions of a stock change more rapidly, a model fitted with past data and patterns is less likely to predict well out-of-sample. For a given stock and day, its volatility on that day is calculated as the standard deviation of its mid-price returns computed at 15 second intervals. Every full trading day is 6.5 hours from 9:30 to 16:00 and is divided into 1560 disjoint 15 second intervals. The jump indicators are binary dummies, indicating whether the absolute values of securities’ daily close-to-close returns fall into the range of 3 − 4%, 4 − 5% or greater than 5%.

Panel regression results are shown in Table 9. We regress the out-of-sample $R ^ { 2 }$ or accuracy measure for the three prediction problems on either volatility only or both volatility and jump indicators. The results show that volatility has a negative impact on the predictability of returns but a positive impact on the predictability of duration, with or without the presence of the jump indicators. Without jump adjustment, one standard deviation increase in volatility reduces, on average, the out-of-sample $R ^ { 2 }$ for 5-second return prediction by 2%. By comparison, volatility has smaller (still negative, but insignificant) impact on the directional accuracy. The presence of jumps (of any of the three sizes) has a significant negative impact on the predictability. Durations, on the other hand, are more predictable for assets with higher volatility and jumps; higher price variability translates into more trading activity, shorter durations, which are more easily predictable.

## D.3 Asset Pricing Characteristics

Next, we examine whether asset pricing characteristics of individual stocks, such as their betas, daily $R ^ { 2 } \mathrm { s }$ , and daily idiosyncratic volatilities, afect the predictability. We first estimate these quantities using a standard first-pass regression. For a given day, let the vector of 15-second mid-price to midprice returns of a stock x be $\mathbf { R } _ { \mathrm { x } }$ . We use SPY (the most liquid exchange traded fund tracking the S&P500 Index) as a proxy for the market portfolio. Then the daily beta of stock x can be calculated as

$$
\operatorname{Beta} (\mathrm{x}, \text { SPY }) = \operatorname{Cov} \left(\mathbf {R} _ {\mathrm{x}}, \mathbf {R} _ {\text { SPY }}\right) / \operatorname{Var} \left(\mathbf {R} _ {\text { SPY }}\right).
$$

Its associated $R ^ { 2 }$ relative to the market portfolio can be calculated as

$$
R ^ {2} (\mathrm{x}, \mathrm{SPY}) = \mathrm{Corr} ^ {2} (\mathbf {R} _ {\mathrm{x}}, \mathbf {R} _ {\mathrm{SPY}}) = \frac {\mathrm{Cov} ^ {2} (\mathbf {R} _ {\mathrm{x}} , \mathbf {R} _ {\mathrm{SPY}})}{\mathrm{Var} (\mathbf {R} _ {\mathrm{x}}) \mathrm{Var} (\mathbf {R} _ {\mathrm{SPY}})},
$$

which measures the percentage of stock x’s movements that can be explained by market movements. Stock x’s own idiosyncratic volatility on that day is calculated as the standard deviation of the residuals of the regression, that is

$$
\mathrm{IdiosyncraticVolatility} = \sqrt {1 - \mathrm{Corr} ^ {2} (\mathbf {R} _ {\mathrm{x}} , \mathbf {R} _ {\mathrm{SPY}})} \mathrm{SD} (\mathbf {R} _ {\mathrm{x}}).
$$

Table 10 summarizes the result of the panel regression. Both Beta and $R ^ { 2 }$ relative to the market portfolio are negatively correlated with return and trade direction predictability, while idiosyncratic volatility follows the same pattern as total volatility in Section D.2. This is consistent with the earlier finding that the returns of more volatile and more liquid stocks are relatively harder to predict. The former is the case for assets with higher betas, which are more volatile due to higher exposure to systematic risk. (Note that we do not attempt to forecast the returns of SPY separately, which would help in this case.) And stocks with higher $R ^ { 2 }$ relative to the market portfolio tend to be more liquid, ceteris paribus.

## D.4 Market-Wide Environment

Turning to the time series determinant of predictability, we now study whether the market portfolio return and its volatility can afect predictability over time. The close-to-close returns (proportional change) of the S&P500 index is used to proxy for market returns. The CBOE VIX index is used to proxy for aggregate market volatility.

The panel regression results are shown in Table 11. We include fixed efects for two sets of dates that are expected to have impact on the equity markets. Those are dates when important economic data are released. The first set contains all the dates of the Federal Open Market Committee (FOMC) meetings, when monetary policy decisions are announced. The second set is all dates when the US Bureau of Labor Statistics releases employment data. A fixed efect will be estimated for every one of these dates: There are 14 FOMC dates and 24 employment data dates from 2019 to 2020, resulting in 38 variables.

As expected, VIX, the market volatility measure, has a negative efect on return and direction predictions. This is in line with the expectation: larger market volatility makes things harder to predict. We also find that days when the market return is positive are more easily predictable than when it is negative, a finding consistent with the classical “leverage efect”. Indeed, the correlation between market returns and volatility is around -0.7.

## E. Additional Results on the Impact of a Delay

Beyond the results presented in Section 4.2, we further examine how a delay afects predictability over time and across the cross-section. First, we assess whether the impact of a delay results in a uniform shift in predictability over time or varies across specific days. Figure 18 replicates the time series results from the second subplot of Figure 16 with a (relatively large) 100 ms delay. The results show that a delay consistently reduces prediction accuracy in the time series, as evidenced by the lower red dotted curve compared to the no-delay blue curve.

To further analyze this efect, Table 13 examines the impact of a delay on days with diferent volatility levels. The daily volatility for a given stock is computed as the standard deviation of its mid-price returns at 15-second intervals, consistent with Section A.4.2. High-volatility days are defined as those with volatility above the stock’s median daily volatility, while low-volatility days are those that fall below this threshold. We find that the impact of a delay is generally more pronounced on low-volatility days. A possible explanation is that high-volatility days tend to exhibit a more persistent signal-to-noise ratio than low-volatility days. Consequently, the noise introduced by a 100 ms delay has a greater efect on low-volatility days, as predictive signals become stale.

Additionally, we examine how this efect varies across stocks with diferent liquidity levels, with the results shown in Table 14. Stocks are categorized by liquidity using both traded volume and proportional spread as measures. Stocks with liquidity measures above the median are classified as having high liquidity, while those below the median are considered having low liquidity. The delay appears to have a slightly larger impact on low-liquidity stocks. A possible reason is similar to the volatility case: stocks with lower liquidity exhibit a less persistent signal-to-noise ratio, making them more susceptible to the noise introduced by delayed updates.

## F. Additional Robustness Checks

## F.1 Fine-tuning the Number of Trees in a Random Forest

Random forests represent the base nonparametric method we used. In the analysis throughout the paper, we fixed the number of trees in each forest at 100. Since each tree is independently and identically distributed, varying the number of trees should only afect the variance of predictions. We report results where the number of trees ranges from 1, 2, 4 to 512. The same procedure for tuning the algorithm is used in each case, as we did for the base case of 100 trees. From the results in Figure 19, we can see that the improvement from increasing the number of trees is limited and we observe almost no diference after 16 trees. An important reason for this is that the performance is averaged over 505 days which is already quite stable. Day-by-day performance however benefits from using additional trees. When more trees are used, tuning selects a deeper depth for each tree as hyper-parameter.

## F.2 Predictability Using Only Subtypes of Data: Trades vs. Quotes

We allowed the algorithms so far to use the merged transaction data and quote update data in all the training and testing. This involves two aspects. First, at each time t both data are used in calculating the predictor variables. Specifically, LobImbalance and QuotedSpread are the only two variables that utilize both data, while other variables are calculated with transaction data only. These two variables can also be calculated with just transaction data. Second, the timestamps t at which we calculate predictor and response variables may come from either a transaction or a quote update event<sup>17</sup>. For INTC in 2019 and 2020, the daily averages of transactions and quote updates are 135K and 630K, respectively. The gap is larger for more liquid assets.

How does limiting the algorithms to using only trade or quote data afect returns predictions? We fit the algorithm with predictors calculated at only timestamps from either trade, quote, or both. To mimic the situation where only transaction data are available, we fit models with only variables calculated with trade data on trade timestamps. Results are presented in Table 15. We can see that the predictability at transaction-only timestamps is significantly higher than those on quote update-only timestamps. This is consistent with the variable selection findings from LASSO in Section 3.1 which isolated transactions-derived variables as the most important predictors of future returns. These diferences might also be related to the diferent temporal distributions of timestamps in each group, or the fact that transaction data are inherently less noisy compared to quotes. And models fitted and tested with the same subtypes of data performed slightly better than models where the subtypes are mixed.

## F.3 Incremental Predictability Using Additional Data From Correlated Stocks

The data employed so far to predict a given stock’s future returns and durations were derived exclusively from observations on that stock’s own past transactions and quotes. We now examine whether additional predictability can be achieved by adding data derived from other stocks. Indeed, correlated stocks do tend to move together but, on a millisecond timescale, correlated moves in diferent stocks are never perfectly synchronized. Whenever a stock might slightly lead another, data on that stock’s transactions and quotes can potentially help predict the laggard stock’s moves.

Continuing with INTC as an example, we add to the set of explanatory variables for INTC predictors that are derived from the four stocks most highly correlated with INTC in terms of daily close-to-close returns from January 2019 to December 2020, among all stocks in the S&P 100. The four stocks selected are TXN, NVDA, MSFT, and QCOM in decreasing correlation order. We then measure the gains in predictability for INTC, if any, by aligning on the basis of their respective timestamps signals from INTC with those from these four stocks.

When using only one additional stock, to keep the number of variables approximately the same, we use the nearest five time spans for each stock $( 5 \times 2 = 1 0$ predictors for each crafted feature), instead of the default nine time spans. Similarly, when recruiting two (resp. four) additional stocks, we use three (resp. two) spans from each in order to keep the number of predictors approximately the same.<sup>18</sup> We compute predictability results using a calendar clock as it makes the most sense when predictors from multiple stocks are merged together.

We find that the use of additional stock information leads only to small improvements. The outof-sample $R ^ { 2 }$ for predicting 5-second INTC returns improves from 18.07% to 18.26% when TXN is added, and further improves to 18.30% when all five stocks are used. The out-of-sample $R ^ { 2 }$ for longer

30 second return improves from 6.40% to 6.62% when including TXN, then 6.78% when NVDA is added and dropped to 6.57% when using 4. The results are shown in Table 16.

## G. Lasso and Random Forests

We now briefly describe the two main methods that we use to predict stock returns and durations: LASSO and random forests (RF).

Consider the regression problem of predicting a response variable Y using a predictor vector X, based on a random sample $\{ ( \mathbf { X } _ { i } , Y _ { i } ) \} _ { i = 1 } ^ { n }$ . Let $\mathbf { Y } = ( y _ { 1 } , y _ { 2 } , \cdot \cdot \cdot , y _ { n } ) ^ { T }$ . In our data, each $\mathbf { X } _ { i }$ has dimension 135, consisting of 9 time spans for each of the 15 predictor variables. The algorithms can then construct to combine these variables for predicting the response variable ${ \cal Y } ,$ as well as select the most informative subsets of variables.

## G.1 Penalized linear regression

One of the simplest methods for predicting response variable Y based on covariates X is the linear model

$$
Y = \boldsymbol {\beta} ^ {T} \mathbf {X} + \varepsilon .
$$

In the absence of some form of regularization, standard OLS in a large dimensional setting is likely to have poor out-of-sample predictive power due to in-sample overfitting. A standard method to address this issue consists of regularizing the model using a penalty function applied to normalized variables. Penalized least-squares with an $L _ { 1 }$ penalty is known as the least absolute shrinkage and selection operator (LASSO). Specifically, let $\begin{array} { r } { \bar { \bf X } = \frac { 1 } { n } \sum _ { i } { \bf X } _ { i } } \end{array}$ and $\begin{array} { r } { s _ { i } = \sqrt { \frac { 1 } { n } \sum _ { i } ( x _ { i } - \bar { x } _ { i } ) ^ { 2 } } ( i = 1 , 2 , \cdot \cdot \cdot , p ) } \end{array}$ be the sample mean vector and the sample standard deviations of predictor variables. Let $\begin{array} { r } { \bar { Y } = \frac { 1 } { n } \sum _ { i = 1 } ^ { n } Y _ { i } } \end{array}$ be the mean of the response variable. Define the centered response $\widetilde { Y _ { i } } = Y _ { i } - \bar { Y }$ and standardized predictors $\widetilde { \mathbf { X } } _ { i } = \operatorname { d i a g } ( s _ { 1 } ^ { - 1 } , s _ { 2 } ^ { - 1 } , \ldots , s _ { p } ^ { - 1 } ) ( { \mathbf { X } } _ { i } - \bar { \mathbf { X } } ) ( i = 1 , \cdots , n )$ . LASSO then fits the centered response on standardized predictors by solving the following optimization problem

$$
\widehat {\pmb {\beta}} = \mathrm{argmin} _ {\pmb {\beta} \in \mathbb {R} ^ {p}} \left\{\frac {1}{n} \sum_ {i} \left(\widetilde {Y} _ {i} - \pmb {\beta} ^ {T} \widetilde {\mathbf {X}} _ {i}\right) ^ {2} + \lambda \| \pmb {\beta} \| _ {1} \right\}.\tag{G.1}
$$

This optimization problem does not admit an analytic solution, but can easily be solved using convex optimization algorithms, coordinate descent methods, least angle regression, among others.<sup>19</sup> For new data $\mathbf { X } _ { \mathrm { n e w } }$ , the model will predict its associated response as

$$
\widehat {Y} _ {\mathrm{new}} = \bar {Y} + \widehat {\boldsymbol {\beta}} ^ {T} \widetilde {X} _ {\mathrm{new}}, \quad \mathrm{with} \quad \widetilde {X} _ {\mathrm{new}} = \mathrm{diag} (s _ {1} ^ {- 1}, s _ {2} ^ {- 1}, \ldots , s _ {p} ^ {- 1}) (\mathbf {X} _ {\mathrm{new}} - \bar {\mathbf {X}}).
$$

LASSO is a simple and useful method that shrinks the coeficients of less useful predictors towards 0. This allows us to rank the relevance of diferent predictors for each prediction problem . <sup>20</sup>

On the other hand, being a least-squares method, LASSO is not robust to heavy-tailed data. As shown in Table 7, the response variables we employ have heavy tails with large kurtosis. To mitigate this problem, we clip our training responses at their $5 ^ { t h }$ and $9 5 ^ { t h }$ percentile to reduce the influence of outliers.<sup>21</sup>

## G.2 Random forests

Classification and regression trees (CART) are important alternatives to parametric methods like LASSO. CART is a scalable nonparametric learning method that can capture the interactions among predictors and nonlinear relationships. A single tree method is known to be unstable and exhibit less predictive power. This leads us to consider ensemble learning trees (see, e.g., Hastie et al. (2009), Murphy (2014), Fan et al. (2020b)). A random forest (RF) prediction is an ensemble estimator of many individual randomized decision trees. By averaging the outcomes of many i.i.d. sampled decision trees via bootstrap samples, the variance of prediction is reduced and the prediction result becomes more stable.

RF is fitted via iteratively growing i.i.d. regression trees by drawing bootstrap samples from the dataset. Then, a decision tree is built by greedily and recursively minimizing the mean squared error (MSE) loss: At each split, a random subset of predictors of size s are considered as the candidate variables for data partition. This makes grown trees from bootstrap samples more independent. Specifically, let N be a bootstrap sample that is used for growing a regression tree. The method first seeks a variable and a location to split the data N. Instead of considering all features, it randomly selects s features with index set $\mathbf { S }$ as candidates. For each feature $k \in \mathbf { S }$ at value x, the method divides the data into two subsets

$$
\mathbf {N} ^ {+} (k, x) = \{i \in \mathbf {N}: X _ {i, k} > x \} \quad \mathrm{and} \quad \mathbf {N} ^ {-} (k, x) = \{i \in \mathbf {N}: X _ {i, k} \leq x \},
$$

where $X _ { i , k }$ is the $k ^ { t h }$ component of the $i ^ { t h }$ sample. Let ${ \bar { Y } } ^ { + }$ be the average of $Y _ { i }$ in $\mathbf { N } ^ { + } ( k , x )$ and $\bar { Y } ^ { - }$ be the average of $Y _ { i }$ in ${ \bf N } ^ { - } ( k , x )$ . The splitting node $( k , x )$ is chosen to minimize the mean square errors:

$$
\widehat {k}, \widehat {x} = \operatorname{argmin} _ {k, x} \left\{\sum_ {i \in \mathbf {N} ^ {+} (k, x)} (Y _ {i} - \bar {Y} _ {k, x} ^ {+}) ^ {2} + \sum_ {i \in \mathbf {N} ^ {-} (k, x)} (Y _ {i} - \bar {Y} _ {k, x} ^ {-}) ^ {2} \right\}
$$

The rest of the tree is grown recursively in this fashion. Both subsets of data $\mathbf { N } ^ { + } ( \widehat { k } , \widehat { x } )$ and $\mathbf { N } ^ { - } ( \widehat { k } , \widehat { x } )$ are partitioned further using a similar method: randomly choosing a subset of variables of size s as the candidate set to partition the data, selecting a variable and a value to optimally split the data. Recursively repeating this splitting process in each subset will yield a regression tree. The stopping criteria include the maximum rounds of splitting (tree depth) and minimum number of data points in each division (leaf size). Iterating the entire process multiple times yields a family of regression trees.

The above process for growing a regression tree divides the sample space into several rectangles and assigns a constant prediction value for each rectangle. Usually the constant is the training sample mean of $Y _ { i }$ in that node of the tree. Specifically, letting $\{ \mathcal { R } _ { j , k } : k \in \mathcal { T } _ { j } \}$ denote the disjoint partitions of the j-th regression tree, the prediction function for the $j ^ { t h }$ regression tree is

$$
\widehat {f} _ {j} (\mathbf {X} _ {\mathrm{new}}) = \sum_ {k \in \mathcal {I} _ {j}} \widehat {\beta} _ {j, k} \mathbb {1} _ {\left\{\mathbf {X} _ {\mathrm{new}} \in \mathcal {R} _ {j, k} \right\}},\tag{G.2}
$$

where $\widehat { \beta } _ { j , k }$ is the sample mean of the responses in training data falling in the partition $\mathcal { R } _ { j , k }$ . In other words, a new predictor $\mathbf { X } _ { \mathrm { n e w } }$ must fall in one of the partitions and the partition average response is used as the prediction by the $j ^ { t h }$ tree. The prediction of a random forest with M trees is the average of their predictions:

$$
\widehat {Y} _ {\mathrm{new}} = \frac {1}{M} \sum_ {j = 1} ^ {M} \widehat {f} _ {j} (\mathbf {X} _ {\mathrm{new}}).
$$

Random forests are improvements over the bagging method where a number of trees are trained independently with bootstrapped samples, but without the process of the random selection of candidate variables at each node. Predictions from bagging are often highly correlated as the bootstrap samples from the same data set tend to overlap. Random splitting in random forests increases the independence of resulting trees and hence reduces the dependence of the prediction. Therefore, the predictions by random forests often end up with a smaller variance than those based on bagging or a single tree, yielding better predictions.

## G.3 Algorithm Tuning and Testing

Each model we employ has a large number of parameters and is tuned and tested on a rolling window basis. A new model is fitted for every testing day using data from the past 5 days, so only the most recent information is included. Hyper-parameters of each method are tuned every month (20 trading days).

Tuning hyper-parameters Hyper-parameters controlling each method need to be tuned periodically in order to accommodate possible structural changes over time. For computational eficiency, we tune only the reduced set of hyper-parameters that matter most to each model. We determine this iteratively. We start with a single stock, INTC (Intel Inc.), to find a range for each parameter. Then we optimize (i.e., tune) over that range separately for each stock. For example, in LASSO we tune the $\ell _ { 1 } { \mathrm { - p e n a l t y } }$ term λ from $1 0 ^ { 6 }$ to $1 0 ^ { - 8 }$ . For RF, we fix the total number of trees at 100, fix the size of each subsample at 100, 0000 (in order to speed up computations) when fitting a single tree. The regression trees are then grown greedily by minimizing the total mean squared error. The depth of each tree is tuned from a range of 3 to 7.

We have also experimented with other choices of hyper-parameters such as the number of days in training a model, a wider range of values of λ, and more depth of trees. The results we obtain are quite robust.

Tuning, training, and testing windows. The rolling window structure is set to ensure that all tests are done with the most up-to-date model and data. As noted, the models used each testing day are fitted with the most recent 5 trading day’s data and all hyperparameters are re-tuned every 20 testing days. All experiments are conducted on a two-layer rolling window basis, corresponding to the training and tuning. An example of one such window is shown in Figure 20. The length of each training window is also set at 5 trading days. The outer layer of rolling window consists of 40 trading days where the first 20 days are used only for tuning hyper-parameters while the next 20 days are used for testing. More specifically, for the current time $T$ that represents a window of 40 trading days $\{ T , T + 1 , \cdots , T + 3 9 \}$ , the procedure is implemented as follows:

1. Learning: For each combination of hyper-parameters and $t = T , T + 5 , T + 1 0$ , fit a model with data from day t to day $t + 4$ (5 trading days), evaluate it on the next 5-day interval $[ t + 5 , t + 9 ]$ and calculate the out-of-sample $R ^ { 2 }$ for each testing day, which results in $R _ { t + 5 } ^ { 2 } , \cdot \cdot \cdot , R _ { t + 9 } ^ { 2 }$

2. Tuning: Choose the combination of hyper-parameters that produces the largest average $R ^ { 2 }$ that is $\textstyle { \frac { 1 } { 1 5 } } \sum _ { t = T + 5 } ^ { T + 1 9 } R _ { t } ^ { 2 }$ , and fix the hyper-parameters for the predictions in the next step.

3. Predicting: For each $t = T + 2 0 , \ldots , T + 3 9$ , fit a model with data from $( t - 5 , \dots , t - 1 )$ and use it to predict on day t. Save each prediction result.

4. Roll forward the entire window by 20 trading days, namely, $T  T + 2 0$ and repeat steps $1 - 4$

Table 7: Summary statistics: TAQ data

<table><tr><td>Data</td><td>mean</td><td>std</td><td>skewness</td><td>kurtosis</td><td>10%</td><td>25%</td><td>50%</td><td>75%</td><td>90%</td></tr><tr><td># data rows</td><td>473K</td><td>217K</td><td>0.7</td><td>-0.3</td><td>224K</td><td>296K</td><td>421K</td><td>625K</td><td>783K</td></tr><tr><td># transactions</td><td>109K</td><td>66K</td><td>1.7</td><td>3.0</td><td>51K</td><td>61K</td><td>84K</td><td>140K</td><td>195K</td></tr><tr><td># quote updates</td><td>364K</td><td>165K</td><td>0.6</td><td>-0.4</td><td>170K</td><td>231K</td><td>325K</td><td>478K</td><td>609K</td></tr><tr><td>Traded Volume (# share)</td><td>10.6M</td><td>16.5M</td><td>5.5</td><td>51.5</td><td>1.7M</td><td>3.0M</td><td>5.5M</td><td>11.1M</td><td>23.2M</td></tr><tr><td>Traded Volume ($)</td><td>1121M</td><td>2488M</td><td>12.2</td><td>381.2</td><td>225M</td><td>338M</td><td>561M</td><td>972M</td><td>1885M</td></tr><tr><td>Total Market Capitalization ($)</td><td>169B</td><td>217B</td><td>4.6</td><td>26.9</td><td>41B</td><td>66B</td><td>111B</td><td>201B</td><td>304B</td></tr><tr><td>Nominal Stock Price ($)</td><td>185.7</td><td>320.3</td><td>5.3</td><td>33.5</td><td>37.5</td><td>54.3</td><td>106.4</td><td>190.5</td><td>317.8</td></tr><tr><td>Daily Return</td><td>0.0</td><td>0.0</td><td>0.1</td><td>16.7</td><td>-0.0</td><td>-0.0</td><td>0.0</td><td>0.0</td><td>0.0</td></tr><tr><td>Daily Beta with SPY</td><td>0.9</td><td>0.3</td><td>0.9</td><td>5.1</td><td>0.5</td><td>0.6</td><td>0.8</td><td>1.0</td><td>1.3</td></tr><tr><td>Daily  $R^2$  with SPY</td><td>0.4</td><td>0.2</td><td>0.0</td><td>-0.6</td><td>0.2</td><td>0.3</td><td>0.4</td><td>0.6</td><td>0.7</td></tr><tr><td>Turnover Rate</td><td>71bp</td><td>96bp</td><td>9.2</td><td>136.0</td><td>27bp</td><td>35bp</td><td>49bp</td><td>74bp</td><td>117bp</td></tr><tr><td>5s returns</td><td>0.0bp</td><td>2.3bp</td><td>0.6</td><td>97.4</td><td>-2.2bp</td><td>-0.9bp</td><td>0.0bp</td><td>0.9bp</td><td>2.1bp</td></tr><tr><td>30s returns</td><td>0.0bp</td><td>4.4bp</td><td>0.1</td><td>39.8</td><td>-4.0bp</td><td>-1.7bp</td><td>0.0bp</td><td>1.7bp</td><td>3.9bp</td></tr><tr><td>10trds returns</td><td>0.0bp</td><td>1.9bp</td><td>1.4</td><td>195.0</td><td>-1.8bp</td><td>-0.9bp</td><td>0.0bp</td><td>0.9bp</td><td>1.8bp</td></tr><tr><td>200trds returns</td><td>0.0bp</td><td>5.9bp</td><td>0.1</td><td>13.6</td><td>-6.1bp</td><td>-2.7bp</td><td>0.0bp</td><td>2.7bp</td><td>6.1bp</td></tr><tr><td>10lots vol returns</td><td>0.0bp</td><td>2.1bp</td><td>0.9</td><td>109.1</td><td>-2.1bp</td><td>-0.9bp</td><td>0.0bp</td><td>0.9bp</td><td>2.1bp</td></tr><tr><td>200lots vol returns</td><td>0.0bp</td><td>7.0bp</td><td>0.0</td><td>18.4</td><td>-6.8bp</td><td>-2.9bp</td><td>0.0bp</td><td>2.9bp</td><td>6.7bp</td></tr><tr><td>10trds duration (seconds)</td><td>7.7</td><td>11.0</td><td>19.6</td><td>3990.8</td><td>0.0</td><td>1.1</td><td>4.2</td><td>10.3</td><td>19.5</td></tr><tr><td>200trds duration (seconds)</td><td>126.3</td><td>137.5</td><td>24.6</td><td>1833.8</td><td>21.6</td><td>46.6</td><td>94.9</td><td>169.7</td><td>266.1</td></tr><tr><td>10lots vol duration (seconds)</td><td>13.0</td><td>68.1</td><td>231.1</td><td>63390.9</td><td>0.1</td><td>1.4</td><td>5.6</td><td>15.0</td><td>31.5</td></tr><tr><td>200lots vol duration (seconds)</td><td>196.8</td><td>305.3</td><td>7.7</td><td>199.0</td><td>16.4</td><td>43.1</td><td>108.0</td><td>234.9</td><td>446.6</td></tr></table>

Note: 1bp = 0.0001 = 0.01%. The upper panel summarizes the variables on aggregated statistics at daily leve across all 101 stocks. It encompasses 101 stocks, 505 trading days from 2019 to 2020 with 50273 samples. The lower panel presents summary statistics of each response variable for each stock and timestamp. The data are downsampled so that each pair (day,stock) contributes a comparable amount of data in calculating the statistics. In the analysis, downsampling is employed solely to accelerate model fitting, with no impact on prediction targe and features. Specifically, we first compute the features and targets using the complete set of transactions and quotes data. During model training, for each rolling-widow, if the training dataset exceeds56 $1 0 ^ { 6 }$ samples, due to th repetitive nature of quotes updates, we randomly select ${ 1 0 } ^ { 6 }$ samples to fit into the model to enhance computational eficiency. Besides computational eficiency, this was also done to avoid specific combinations of (day,stock) having too high a weight in the training. Overall, the downsampling results in removing 3.3% of the data, which breaks

Figure 16: Average daily performance of RF for transaction return, direction and duration predictions across S&P 100 stocks  
![](images/26c7b625f4149d541554c5156df36412053b0c6e0f8bee5bb80b54a406a74299.jpg)

![](images/3d95eee302361f1decadfb970d9b38dab69387d8087954837928347bb7cc3cb8.jpg)

![](images/97d946fcb49cd3758b5de43e5feb882bb4bb870b904cbd719cb7e5e9f8142c2b.jpg)

![](images/1b01c9a4986f9acbd180e6a9c40299f9565b4ab1e3d9b5d2eeef798b399add55.jpg)  
Note: The top panel shows the S&P 100 index along with its daily range as a proxy for its variability. The bottom three panels depict daily average out-of-sample $R ^ { 2 } { \mathrm { s } }$ (blue curve) and their associated standard deviations (shaded bars), over stocks, for predicting 5-second returns, their signs, and 10-trade duration. RF is used for prediction. The horizontal black lines indicate the performance resulting from using the in-sample average (second and fourth panel) or random guesses (third panel).

Figure 17: Prediction performance as a function of liquidity: Transaction volume and bid-ask spread  
![](images/2998f862f606bd19e0ed8edd658eebfdb61e16fd5dd5128259341e6f25bf5c0b.jpg)  
Note: Each point represents a security’s average daily performance and average daily dollar traded volume / proportional spread in 2019 and 2020. The same response variables as in Figure 13 are used. The red lines are the OLS fits of data. The proportional spread each day is calculated as the security’s average bid-ask spreads divided by mid prices, sampled every 15 seconds.

Table 8: Regression of predictability on liquidity measures

<table><tr><td></td><td>5s return</td><td>5s direction</td><td>10 txn duration</td></tr><tr><td>Traded Volume (log)</td><td>-0.019***(0.001)</td><td>-0.002***(0.0003)</td><td>0.057***(0.001)</td></tr><tr><td>Proportional Spread</td><td>0.005***(0.0003)</td><td>0.005***(0.0002)</td><td>0.003***(0.0004)</td></tr><tr><td>Symbol fixed effect</td><td>YES</td><td>YES</td><td>YES</td></tr><tr><td>Date fixed effect</td><td>YES</td><td>YES</td><td>YES</td></tr><tr><td>Adjusted R2</td><td>0.706</td><td>0.691</td><td>0.228</td></tr></table>

Notation: <sup>∗</sup>p<0.1; <sup>∗∗</sup>p<0.05; <sup>∗∗∗</sup>p<0.01

Table 9: Regression of predictability on volatility and jump measures

<table><tr><td></td><td colspan="2">5s return</td><td colspan="2">5s direction</td><td colspan="2">10 txn duration</td></tr><tr><td>Volatility</td><td>-0.019***(0.001)</td><td>-0.016***(0.001)</td><td>-0.001**(0.0003)</td><td>-0.0002(0.0003)</td><td>0.043***(0.001)</td><td>0.032***(0.001)</td></tr><tr><td>Jump 3-4%</td><td></td><td>-0.014***(0.001)</td><td></td><td>-0.0003(0.001)</td><td></td><td>0.026***(0.002)</td></tr><tr><td>Jump 4-5%</td><td></td><td>-0.014***(0.001)</td><td></td><td>-0.002*(0.001)</td><td></td><td>0.051***(0.002)</td></tr><tr><td>Jump 5%+</td><td></td><td>-0.022***(0.001)</td><td></td><td>-0.004***(0.001)</td><td></td><td>0.074***(0.002)</td></tr><tr><td>Symbol fixed effect</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td></tr><tr><td>Date fixed effect</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td></tr><tr><td>Adjusted R $^{2}$ </td><td>0.704</td><td>0.707</td><td>0.684</td><td>0.685</td><td>0.193</td><td>0.220</td></tr></table>

Notation: <sup>∗</sup>p<0.1; <sup>∗∗</sup>p<0.05; <sup>∗∗∗</sup>p<0.01

Table 10: Regression of predictability on asset pricing characteristics

<table><tr><td></td><td>5s return</td><td>5s direction</td><td>10 txn duration</td></tr><tr><td>Beta with SPY</td><td>-0.007***(0.001)</td><td>-0.001**(0.0003)</td><td>0.016***(0.001)</td></tr><tr><td> $R^2$  with SPY</td><td>-0.008***(0.001)</td><td>-0.012***(0.0005)</td><td>-0.035***(0.001)</td></tr><tr><td>Idiosyncratic Volatility</td><td>-0.011***(0.001)</td><td>-0.001***(0.0003)</td><td>0.016***(0.001)</td></tr><tr><td>Symbol fixed effect</td><td>YES</td><td>YES</td><td>YES</td></tr><tr><td>Date fixed effect</td><td>YES</td><td>YES</td><td>YES</td></tr><tr><td>Adjusted  $R^2$ </td><td>0.708</td><td>0.698</td><td>0.214</td></tr></table>

Notation: <sup>∗</sup>p<0.1; <sup>∗∗</sup>p<0.05; <sup>∗∗∗</sup>p<0.01

Table 11: Regression of predictability on aggregate market conditions

<table><tr><td></td><td colspan="2">5s return</td><td colspan="2">5s direction</td><td colspan="2">10 txn duration</td></tr><tr><td>VIX</td><td>-0.007***(0.0002)</td><td></td><td>-0.012***(0.0001)</td><td></td><td>-0.003***(0.0004)</td><td></td></tr><tr><td>S&amp;P500 Return</td><td></td><td>0.003***(0.0002)</td><td></td><td>0.002***(0.0001)</td><td></td><td>0.002***(0.0004)</td></tr><tr><td>Symbol fixed effect</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td></tr><tr><td>Date fixed effect</td><td>NO</td><td>NO</td><td>NO</td><td>NO</td><td>NO</td><td>NO</td></tr><tr><td>Data Date fixed effect</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td></tr><tr><td>Adjusted R $^{2}$ </td><td>0.616</td><td>0.611</td><td>0.644</td><td>0.588</td><td>0.041</td><td>0.040</td></tr></table>

Notation: <sup>∗</sup>p<0.1; <sup>∗∗</sup>p<0.05; <sup>∗∗∗</sup>p<0.01

Figure 18: Average daily performance of RF for transaction return predictions across S&P 100 stocks: with and without DelayAverage Return R^  
![](images/1dadf3bc07a591b341a6cba902a6cff3d2f2dbec2e2969af1667a7b103709ff4.jpg)  
Note: Daily average out-of-sample $R ^ { 2 } \mathrm { s }$ (curves) and their associated standard deviations (shaded bars), over stocks, for predicting 5-second returns. RF is used for prediction. The blue curve with gray shade is the result without delay (same as the second subfigure in Figure 16), and the red dotted curve with red shade is the result with a 100 ms delay. The horizontal black lines indicate the performance resulting from using the in-sample average.

Table 12: Determinants of predictability for duration: Panel regression of out-of-sample $R ^ { 2 }$

<table><tr><td rowspan="2"></td><td colspan="8">Dependent variable:  $R^2$ s of Duration Predictions</td></tr><tr><td>10 txn</td><td>10 txn</td><td>200 txn</td><td>200 txn</td><td>1K vol</td><td>1K vol</td><td>20K vol</td><td>20K vol</td></tr><tr><td>Nominal Share Price (log)</td><td>0.007**(0.003)</td><td>0.005(0.004)</td><td>0.010(0.008)</td><td>0.004(0.008)</td><td>-0.003(0.004)</td><td>-0.004(0.004)</td><td>-0.015*(0.009)</td><td>-0.022**(0.009)</td></tr><tr><td>Daily Return (log)</td><td>-0.004***(0.0004)</td><td>-0.004***(0.0004)</td><td>-0.002**(0.001)</td><td>-0.002*(0.001)</td><td>-0.002***(0.0005)</td><td>-0.002***(0.0005)</td><td>-0.002*(0.001)</td><td>-0.002*(0.001)</td></tr><tr><td>Traded Volume (log)</td><td>0.034***(0.001)</td><td>0.041***(0.001)</td><td>0.003(0.003)</td><td>0.018***(0.002)</td><td>0.036***(0.001)</td><td>0.040***(0.001)</td><td>0.009***(0.003)</td><td>0.019***(0.003)</td></tr><tr><td>Total Market Cap (log)</td><td>0.0003(0.003)</td><td>-0.003(0.003)</td><td>0.022***(0.007)</td><td>0.008(0.007)</td><td>0.003(0.003)</td><td>0.0002(0.003)</td><td>0.024***(0.007)</td><td>0.017**(0.007)</td></tr><tr><td>Proportional Spread</td><td>-0.001**(0.0004)</td><td>-0.005***(0.0004)</td><td>-0.008***(0.001)</td><td>-0.014***(0.001)</td><td>-0.001***(0.001)</td><td>-0.004***(0.0004)</td><td>-0.003**(0.001)</td><td>-0.010***(0.001)</td></tr><tr><td>Volatility</td><td>0.006***(0.001)</td><td>-0.004***(0.001)</td><td>0.011***(0.003)</td><td>0.005**(0.002)</td><td>0.007***(0.002)</td><td>0.008***(0.001)</td><td>0.018***(0.003)</td><td>0.015***(0.002)</td></tr><tr><td>Beta with SPY</td><td>-0.003***(0.001)</td><td>-0.0002(0.001)</td><td>0.004(0.003)</td><td>0.015***(0.001)</td><td>-0.003**(0.001)</td><td>-0.0002(0.001)</td><td>0.004(0.003)</td><td>0.012***(0.002)</td></tr><tr><td> $R^2$  with SPY</td><td>-0.001(0.002)</td><td>-0.002**(0.001)</td><td>-0.007(0.004)</td><td>-0.019***(0.002)</td><td>-0.005**(0.002)</td><td>-0.010***(0.001)</td><td>-0.018***(0.005)</td><td>-0.030***(0.002)</td></tr><tr><td>Idiosyncratic Volatility</td><td>0.020***(0.002)</td><td>0.020***(0.001)</td><td>0.023***(0.004)</td><td>0.012***(0.003)</td><td>0.016***(0.002)</td><td>0.010***(0.001)</td><td>0.009**(0.004)</td><td>0.0004(0.003)</td></tr><tr><td>Jump 3-4 %</td><td>0.017***(0.002)</td><td>0.016***(0.002)</td><td>0.022***(0.004)</td><td>0.016***(0.004)</td><td>0.011***(0.002)</td><td>0.009***(0.002)</td><td>0.003(0.004)</td><td>-0.005(0.004)</td></tr><tr><td>Jump 4-5 %</td><td>0.038***(0.002)</td><td>0.035***(0.002)</td><td>0.052***(0.005)</td><td>0.044***(0.005)</td><td>0.031***(0.002)</td><td>0.026***(0.002)</td><td>0.038***(0.006)</td><td>0.025***(0.005)</td></tr><tr><td>Jump 5 %+</td><td>0.060***(0.002)</td><td>0.040***(0.002)</td><td>0.063***(0.005)</td><td>0.038***(0.005)</td><td>0.048***(0.002)</td><td>0.035***(0.002)</td><td>0.048***(0.005)</td><td>0.024***(0.005)</td></tr><tr><td>S&amp;P500 Return</td><td></td><td>0.004***(0.0004)</td><td></td><td>0.002(0.001)</td><td></td><td>-0.006***(0.001)</td><td></td><td>-0.012***(0.001)</td></tr><tr><td>VIX</td><td></td><td>-0.023***(0.001)</td><td></td><td>-0.023***(0.002)</td><td></td><td>-0.027***(0.001)</td><td></td><td>-0.029***(0.002)</td></tr><tr><td>Symbol fixed effect</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td><td>YES</td></tr><tr><td>Date fixed effect</td><td>YES</td><td>NO</td><td>YES</td><td>NO</td><td>YES</td><td>NO</td><td>YES</td><td>NO</td></tr><tr><td>Data Date fixed effect</td><td>NO</td><td>YES</td><td>NO</td><td>YES</td><td>NO</td><td>YES</td><td>NO</td><td>YES</td></tr><tr><td>Adjusted  $R^2$ </td><td>0.266</td><td>0.182</td><td>0.164</td><td>0.099</td><td>0.216</td><td>0.173</td><td>0.126</td><td>0.090</td></tr></table>

Notation: <sup>∗</sup>p<0.1; <sup>∗∗</sup>p<0.05; <sup>∗∗∗</sup>p<0.01

Note: The response variables are out-of-sample $R _ { i , t } ^ { 2 } \mathrm { s }$ of duration predictions for each security i and date t. Fixed efects for symbols control each i and dates control each t. Data dates fixed efects include fixed efects on dates when important market data are released, which include FOMC release dates and monthly unemployment data release dates. We regressed over all securities in S&P100 at the end of 2020. Each explanatory variables are standardized before regression, except the three jump indicators.

Table 13: Impact of a delay on the predictability over time: Out-of-sample $R ^ { 2 }$ as a function of day’s volatility

<table><tr><td></td><td>Full Sample</td><td>High Volatility Days</td><td>Low Volatility Days</td></tr><tr><td>100 ms delay</td><td>0.020 (0.038)</td><td>0.022 (0.040)</td><td>0.018 (0.034)</td></tr><tr><td>No delay</td><td>0.120 (0.085)</td><td>0.110 (0.086)</td><td>0.128 (0.083)</td></tr></table>

Note: Average (over days and stocks) out-of-sample $R ^ { 2 }$ (and standard deviations) of 5-second return predictions with 100 ms delay (first row) or without delay (second row). Results are reported for the full sample, high-volatility days, and low-volatility days, respectively.

Table 14: Impact of delay on the predictability in the cross-section: Out-of-sample $R ^ { 2 }$ as a function of stocks’ liquidity

<table><tr><td></td><td></td><td>Full Sample</td><td>High Liquidity Stocks</td><td>Low Liquidity Stocks</td></tr><tr><td rowspan="2">2*traded volume</td><td>100 ms delay</td><td>0.020 (0.032)</td><td>0.021 (0.030)</td><td>0.019 (0.034)</td></tr><tr><td>no delay</td><td>0.119 (0.066)</td><td>0.113 (0.065)</td><td>0.125 (0.065)</td></tr><tr><td rowspan="2">2*proportional spread</td><td>100 ms delay</td><td>0.020 (0.032)</td><td>0.013 (0.014)</td><td>0.026 (0.043)</td></tr><tr><td>no delay</td><td>0.119 (0.066)</td><td>0.097 (0.045)</td><td>0.140 (0.076)</td></tr></table>

Note: Average out-of-sample $R ^ { 2 }$ (along with standard deviations) for 5-second return predictions with a 100 ms delay, computed across days and stocks. Results are reported for the full sample, high-liquidity stocks, and lowliquidity stocks. Stock liquidity is assessed using two distinct measures: traded volume and proportional spread

Figure 19: Sensitivity of predictability performance to the number of trees in random forests  
![](images/476ecd040b9334bc2346bc32edcd5196ce578bac2ae7e64bfa75567bd2a2df2a.jpg)

![](images/622b447402afbcfe3371672d96e5077ef268ae40b8cd73816f9ba74162277a0d.jpg)  
Note: Left and right panels are performance in return and direction predictions respectively. The shaded area depicts 95% confidence intervals of the mean out-of-sample $R ^ { 2 }$ or accuracy over 505 days. Black lines are the baseline performances for each problem. The baseline for return $R ^ { 2 }$ is the sample mean of the training samples, and for accuracy is a random guess.

Table 15: Predictability of returns using subtypes of data: Transactions-only vs. quotes-only data

<table><tr><td rowspan="2">Training Timestamps</td><td rowspan="2">Used Data</td><td colspan="3">Testing Timestamps</td></tr><tr><td>Trade&#x27;s</td><td>Quote&#x27;s</td><td>Both</td></tr><tr><td>Trade&#x27;s</td><td>Trade</td><td>18.0% (8.2%)</td><td>12.4% (7.1%)</td><td>13.4% (7.3%)</td></tr><tr><td>Quote&#x27;s</td><td>Both</td><td>16.8% (8.1%)</td><td>13.3% (7.4%)</td><td>13.9% (7.4%)</td></tr><tr><td>Both</td><td>Both</td><td>17.3% (8.1%)</td><td>13.3% (7.4%)</td><td>14.0% (7.5%)</td></tr></table>

Note: Average out-of-sample $R ^ { 2 }$ (and standard deviations) of 5 second return predictions with trade or quote data. Each row uses timestamps and data from a specific subtype of data to calculate all the predictor variables in both fitting and predictions. Each column uses a diferent group of timestamps in testing.

Table 16: Predictability using data from additional stocks

<table><tr><td>Total #stocks</td><td>Additional symbols</td><td>#spans</td><td>5s return  $R^{2}$ </td><td>30s return  $R^{2}$ </td></tr><tr><td>1</td><td></td><td>9</td><td>18.07% (18.50%)</td><td>6.40% (6.45%)</td></tr><tr><td>2</td><td>TXN</td><td>5</td><td>18.26% (18.79%)</td><td>6.62% (6.46%)</td></tr><tr><td>3</td><td>TXN, NVDA</td><td>3</td><td>18.30% (18.79%)</td><td>6.78% (6.56%)</td></tr><tr><td>5</td><td>TXN, NVDA, MSFT, QCOM</td><td>2</td><td>18.30% (18.69%)</td><td>6.57% (6.42%)</td></tr></table>

Note: Average out-of-sample $R ^ { 2 }$ (and standard deviations) of 5-second return predictions and 30-second return predictions. The signals from mostly correlated stocks are merged into INTC as of the newest signal at or before each data point in INTC based on their respective timestamps. For each crafted feature, the number of predictors with merged stocks is (total number of stocks) × (number of spans). The number of spans is chosen so that the total number of predictors is approximately the same, 9 or 10.

Figure 20: Algorithm tuning and testing rolling windows

$$
\begin{array}{c c c} \text {tune 1: 5d train} & \text {5d test} \\ \text {tune 2: 5d train} & \text {5d test} \\ \text {tune 3: 5d train} & \text {5d test} \\ \text {time} & \text {test 1: 5d train} & \text {1d test} \\ & \text {test 20: 5d train} & \text {1d test} \\ \text {Tuning Period (20 days): T, ..., T+19} & & \text {Testing Period (20 days): T+20, ..., T+39} \\ \hline \end{array}
$$

Note: The tuning parameters are optimized once every 20 trading days. For each given set of tuning parameters, we use the past 5 trading data to train the model and the next five days data to compute the testing errors. Thi is done once every 5 days (illustrated above the black line). The testing errors in the last 15 days (orange color) are used to choose the optimal set of tuning parameters. With the selected tuning parameters, we use past 5 days data to predict the next day target (colored red) in a rolling window manner (indicated above the green line) for the next twenty days. The cycle process repeats once every 20 days.