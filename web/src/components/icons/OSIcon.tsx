import Android from "~/components/icons/Android";
import Windows from "~/components/icons/Windows";
import Linux from "~/components/icons/Linux";
import { cx, isAndroid, isApple, isLinux, isWindows } from "~/lib/utils";
import { Match, Switch } from "solid-js";
import Apple from "./Apple";

interface Props {
	os: string;
	class?: string;
}

export default function OSIcon(props: Props) {
	const iconClass = cx("size-5 text-zinc-400", props.class);

	return (
		<Switch>
			<Match when={isApple(props.os)}>
				<Apple class={iconClass} />
			</Match>
			<Match when={isAndroid(props.os)}>
				<Android class={iconClass} />
			</Match>
			<Match when={isLinux(props.os)}>
				<Linux class={iconClass} />
			</Match>
			<Match when={isWindows(props.os)}>
				<Windows class={iconClass} />
			</Match>
		</Switch>
	);
}
