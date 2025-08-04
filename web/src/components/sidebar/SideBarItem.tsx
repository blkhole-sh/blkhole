import { A, useLocation } from "@solidjs/router";

interface Props {
	icon: string;
	title: string;
	url: string;
}

export default function SideBarItem(props: Props) {
	const location = useLocation();

	const selected = () => location.pathname === props.url;

	const styles = () => ({
		"bg-gray-50": selected(),
		"hover:bg-gray-50": !selected(),
	});

	return (
		<li>
			<ul role="list" class="space-y-1">
				<li>
					<A
						href={props.url}
						class="group flex flex-row gap-x-4 items-center p-2 text-sm text-gray-600 font-medium leading-6"
						classList={styles()}
					>
						<span class="text-lg">{props.icon}</span>
						{props.title}
					</A>
				</li>
			</ul>
		</li>
	);
}
