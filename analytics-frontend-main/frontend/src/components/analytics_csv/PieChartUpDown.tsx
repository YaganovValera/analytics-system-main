// src/components/PieChartUpDown.tsx
import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer } from 'recharts';
import './PieChartUpDown.css';

interface Props {
  upRatio: number;
  downRatio: number;
}

const COLORS = ['#4caf50', '#f44336'];
const LABELS = ['Рост (вверх)', 'Падение (вниз)'];

function PieChartUpDown({ upRatio, downRatio }: Props) {
  const data = [
    { name: LABELS[0], value: upRatio },
    { name: LABELS[1], value: downRatio },
  ];

  return (
    <div className="pie-chart-block">
      <h3> Соотношение свечей роста и падения</h3>
      <ResponsiveContainer width="100%" height={300}>
        <PieChart>
          <Pie
            data={data}
            cx="50%"
            cy="50%"
            labelLine={false}
            label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(1)}%`}
            outerRadius={100}
            dataKey="value"
          >
            {data.map((_, index) => (
              <Cell key={`cell-${index}`} fill={COLORS[index]} />
            ))}
          </Pie>
          <Tooltip formatter={(value: number) => `${(value * 100).toFixed(2)}%`} />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
}

export default PieChartUpDown;