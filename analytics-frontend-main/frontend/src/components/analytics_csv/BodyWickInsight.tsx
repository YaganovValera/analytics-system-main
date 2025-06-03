// src/components/BodyWickInsight.tsx
import { useState } from 'react';
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, Label } from 'recharts';
import type { AnalyticsResponse } from '../../types/analytics';
import './BodyWickInsight.css';

interface Props {
    analytics: AnalyticsResponse['analytics'];
}

function BodyWickInsight({ analytics }: Props) {
    const [flipped, setFlipped] = useState(false);

    const { avg_body_size, avg_upper_wick, avg_lower_wick } = analytics;
    const totalWick = avg_upper_wick + avg_lower_wick;
    const bodyRatio = totalWick > 0 ? avg_body_size / totalWick : 0;
    const wickRatio = avg_body_size > 0 ? totalWick / avg_body_size : 0;

    const data = [
        { name: 'Тело свечи', value: avg_body_size },
        { name: 'Верхняя тень', value: avg_upper_wick },
        { name: 'Нижняя тень', value: avg_lower_wick },
    ];

    const verdict = avg_body_size > totalWick
        ? `Среднее тело свечи составляет ${avg_body_size.toFixed(2)}, что в ${bodyRatio.toFixed(1)} раза больше суммы теней (${totalWick.toFixed(2)}). Это указывает на устойчивое и уверенное движение рынка.`
        : totalWick > avg_body_size * 1.5
        ? `Сумма теней (${totalWick.toFixed(2)}) в ${wickRatio.toFixed(1)} раза превышает тело свечи (${avg_body_size.toFixed(2)}). Это говорит о неуверенности и колебаниях на рынке.`
        : `Тело свечи (${avg_body_size.toFixed(2)}) и тени (${totalWick.toFixed(2)}) примерно равны. Это может означать боковое или смешанное движение.`;

    return (
        <div className="wick-insight-wrapper">
        <div className="wick-chart">
            <h4>🕯️ Структура свечей (средние значения)</h4>
            <ResponsiveContainer width="100%" height={260}>
            <LineChart data={data} margin={{ top: 20, right: 30, left: 10, bottom: 10 }}>
                <XAxis dataKey="name" />
                <YAxis>
                <Label
                    angle={-90}
                    position="insideLeft"
                    style={{ textAnchor: 'middle', fill: '#666' }}
                >
                    Значение (в пунктах)
                </Label>
                </YAxis>
                <Tooltip formatter={(v: number) => v.toFixed(2)} />
                <Line
                type="monotone"
                dataKey="value"
                stroke="#1976d2"
                strokeWidth={3}
                dot={{ r: 6 }}
                activeDot={{ r: 8 }}
                />
            </LineChart>
            </ResponsiveContainer>
        </div>
        <div className={`wick-summary ${flipped ? 'flipped' : ''}`}> 
            <div className="flip-icon" onClick={() => setFlipped(!flipped)} title={flipped ? 'Назад к анализу' : 'Показать инструкцию'}>
            {flipped ? '🔙' : 'ℹ️'}
            </div>
            <div className="wick-text">
            {!flipped ? (
                <>
                <h4>📊 Автоматический анализ</h4>
                <p><strong>Среднее тело:</strong> {avg_body_size.toFixed(2)}</p>
                <p><strong>Верхняя тень:</strong> {avg_upper_wick.toFixed(2)}</p>
                <p><strong>Нижняя тень:</strong> {avg_lower_wick.toFixed(2)}</p>
                <p>{verdict}</p>
                </>
            ) : (
                <>
                <h4>📘 Что это значит?</h4>
                <p>Каждая свеча состоит из тела и теней. Тело — это диапазон между ценами открытия и закрытия. Тени — максимумы и минимумы, в которые цена сходила.</p>
                <p>Если тело больше теней — рынок двигался уверенно в одном направлении. Если тени длиннее тела — было много колебаний, неуверенность, борьба между покупателями и продавцами.</p>
                <p>Анализ средней структуры свечей помогает понять, насколько «спокойным» или «турбулентным» был рынок в периоде анализа.</p>
                </>
            )}
            </div>
        </div>
        </div>
    );
}

export default BodyWickInsight;