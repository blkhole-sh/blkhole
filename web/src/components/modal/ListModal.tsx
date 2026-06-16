import { createEffect, createSignal, Show } from "solid-js";
import { createList, updateList } from "~/lib/api";
import type { List } from "~/lib/model";
import { compose, url as isUrl, required } from "~/lib/validate";
import FileInput from "../form/FileInput";
import TextAreaInput from "../form/TextAreaInput";
import TextInput from "../form/TextInput";
import TabBar from "../ui/TabBar";
import Modal from "./Modal";

type SourceTab = "url" | "file" | "manual";

const SOURCE_TABS = [
	{ key: "manual" as SourceTab, label: "MANUAL ENTRY" },
	{ key: "file" as SourceTab, label: "FILE UPLOAD" },
	{ key: "url" as SourceTab, label: "URL" },
];

interface Props {
	open: boolean;
	list?: List | null;
	onClose: () => void;
	onSaved: () => void;
}

export default function ListModal(props: Props) {
	const [name, setName] = createSignal("");
	const [description, setDescription] = createSignal("");
	const [urlValue, setUrlValue] = createSignal("");
	const [manualValue, setManualValue] = createSignal("");
	const [source, setSource] = createSignal("");
	const [sourceTab, setSourceTab] = createSignal<SourceTab>("manual");
	const [submitted, setSubmitted] = createSignal(false);
	const [error, setError] = createSignal("");

	const isEditMode = () => !!props.list;
	const title = () => (isEditMode() ? "Edit Blocklist" : "Create Blocklist");

	// Pre-fill form when opening in edit mode
	createEffect(() => {
		if (props.list && props.open) {
			setName(props.list.name);
			setDescription(props.list.description);
			setSource(props.list.source);
			// Determine source tab from source content
			if (
				props.list.source.startsWith("http://") ||
				props.list.source.startsWith("https://")
			) {
				setSourceTab("url");
				setUrlValue(props.list.source);
			} else {
				setSourceTab("manual");
				setManualValue(props.list.source);
			}
		} else if (!props.open) {
			// Reset when modal closes
			setName("");
			setDescription("");
			setUrlValue("");
			setManualValue("");
			setSource("");
			setSourceTab("manual");
			setSubmitted(false);
			setError("");
		}
	});

	const handleClose = () => {
		setSubmitted(false);
		setName("");
		setDescription("");
		setUrlValue("");
		setManualValue("");
		setSource("");
		setSourceTab("manual");
		setError("");
		props.onClose();
	};

	const handleSubmit = async () => {
		setSubmitted(true);
		if (!name().trim() || !source().trim()) return;
		try {
			setError("");
			const list = props.list;
			if (list) {
				await updateList(list.id, name(), description(), source());
			} else {
				await createList(name(), description(), source());
			}
			setName("");
			setDescription("");
			setUrlValue("");
			setManualValue("");
			setSource("");
			setSourceTab("manual");
			setSubmitted(false);
			props.onSaved();
		} catch {
			setError(`Failed to ${isEditMode() ? "update" : "create"} blocklist.`);
		}
	};

	return (
		<Modal
			title={title()}
			open={props.open}
			onClose={handleClose}
			onConfirm={handleSubmit}
		>
			<div class="flex flex-col gap-8">
				<Show when={!isEditMode()}>
					<TabBar
						tabs={SOURCE_TABS}
						active={sourceTab()}
						onChange={setSourceTab}
					/>
				</Show>
				<TextInput
					label="BLOCKLIST NAME"
					placeholder="Ads & Trackers"
					value={name()}
					onInput={(e) => setName(e.currentTarget.value)}
					validate={required()}
					showError={submitted()}
				/>
				<TextAreaInput
					label="DESCRIPTION"
					placeholder="Blocks ads and tracking scripts."
					value={description()}
					class="min-h-24 max-h-64"
					onInput={(e) => setDescription(e.currentTarget.value)}
					showError={submitted()}
				/>
				<Show when={sourceTab() === "url"}>
					<TextInput
						label="URL"
						placeholder="https://raw.githubusercontent.com/..."
						hint="Supports hosts files, Adblock Plus filter lists, and plain domain lists."
						value={urlValue()}
						onInput={(e) => {
							setUrlValue(e.currentTarget.value);
							setSource(e.currentTarget.value);
						}}
						validate={compose(required(), isUrl())}
						showError={submitted()}
					/>
				</Show>
				<Show when={sourceTab() === "file"}>
					<FileInput
						label="FILE"
						accept=".txt,.csv,.list"
						hint="Supports hosts files, Adblock Plus filter lists, and plain domain lists."
						onChange={(content) => setSource(content)}
						showError={submitted()}
					/>
				</Show>
				<Show when={sourceTab() === "manual"}>
					<TextAreaInput
						label="DOMAINS"
						placeholder={"example.com\nads.example.com"}
						value={manualValue()}
						class="min-h-32"
						hint="One domain per line."
						onInput={(e) => {
							setManualValue(e.currentTarget.value);
							setSource(e.currentTarget.value);
						}}
						validate={required()}
						showError={submitted()}
					/>
				</Show>
			</div>
			{error() && <p class="text-sm text-red-700">{error()}</p>}
		</Modal>
	);
}
