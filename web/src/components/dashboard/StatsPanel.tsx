import type { QueryStats } from "~/lib/model";

interface Props {
	stats: QueryStats | undefined;
}

export default function StatsCards(props: Props) {
	// Total queries, or "-" if none
	const totalQueries = () => {
		if (!props.stats) return "-";
		const total = props.stats.total.reduce((sum, s) => sum + s.count, 0);
		return total === 0 ? "-" : total.toLocaleString();
	};

	// Block rate as percentage, or "-" if none
	const blockRate = () => {
		if (!props.stats) return "-";
		const total = props.stats.total.reduce((sum, s) => sum + s.count, 0);
		const blocked = props.stats.blocked.reduce((sum, s) => sum + s.count, 0);
		if (total === 0) return "-";
		return `${Math.round((blocked / total) * 100)}%`;
	};

	return (
		<div class="pb-8 grid grid-cols-2 items-start gap-8">
			<div>
				<p class="pb-4 font-medium text-zinc-700 text-sm tracking-wider">
					TOTAL QUERIES
				</p>
				<p class="font-display text-5xl">{totalQueries()}</p>
			</div>
			<div>
				<p class="pb-4 font-medium text-zinc-700 text-sm tracking-wider">
					BLOCK RATE
				</p>
				<p class="font-display text-5xl">{blockRate()}</p>
			</div>
		</div>
	);
}
