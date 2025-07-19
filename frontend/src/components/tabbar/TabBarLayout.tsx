import { ParentProps } from "solid-js";
import TabBar, { Tab } from "./TabBar";

interface Props {
	tabs: Tab[];
}

export function TabBarLayout(props: Props & ParentProps) {
	return (
		<div class="font-inter flex flex-col h-screen w-full overflow-hidden py-6 px-8">
			<TabBar tabs={props.tabs} />
			<div class="flex-1 overflow-hidden">{props.children}</div>
		</div>
	);
}
