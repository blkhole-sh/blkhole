import { createEffect, createSignal, For, Show } from "solid-js";
import { createList, updateList } from "~/lib/api";
import type { List } from "~/lib/model";
import {
	compose,
	domain as isDomain,
	url as isUrl,
	required,
} from "~/lib/validate";
import FileInput from "../form/FileInput";
import TextInput from "../form/TextInput";
import ActionButton from "../ui/ActionButton";
import Divider from "../ui/Divider";
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
	const [urlValue, setUrlValue] = createSignal("");
	const [fileValue, setFileValue] = createSignal("");
	const [manualDraft, setManualDraft] = createSignal("");
	const [manualDomains, setManualDomains] = createSignal<string[]>([]);
	const [sourceTab, setSourceTab] = createSignal<SourceTab>("manual");
	const [submitted, setSubmitted] = createSignal(false);
	const [error, setError] = createSignal("");
	const [manualError, setManualError] = createSignal("");

	const isEditMode = () => !!props.list;
	const title = () => (isEditMode() ? "Edit Blocklist" : "Create Blocklist");
	const sourceValue = () => {
		switch (sourceTab()) {
			case "url":
				return urlValue();
			case "file":
				return fileValue();
			default:
				return manualDomains().join("\n");
		}
	};

	const reset = () => {
		setName("");
		setUrlValue("");
		setFileValue("");
		setManualDraft("");
		setManualDomains([]);
		setSourceTab("manual");
		setSubmitted(false);
		setError("");
		setManualError("");
	};

	const addManualDomain = () => {
		const value = manualDraft().trim().toLowerCase();
		const validationError = isDomain()(value);
		if (validationError) {
			setManualError(validationError);
			return;
		}
		if (manualDomains().includes(value)) {
			setManualError("Domain already added");
			return;
		}
		setManualDomains((domains) => [...domains, value]);
		setManualDraft("");
		setManualError("");
	};

	// Pre-fill form when opening in edit mode
	createEffect(() => {
		if (props.list && props.open) {
			setName(props.list.name);
			if (
				props.list.source.startsWith("http://") ||
				props.list.source.startsWith("https://")
			) {
				setSourceTab("url");
				setUrlValue(props.list.source);
			} else {
				setSourceTab("manual");
				setManualDomains(
					props.list.source
						.split(/\r?\n/)
						.map((value) => value.trim())
						.filter(Boolean),
				);
			}
		} else if (!props.open) {
			reset();
		}
	});

	const handleClose = () => {
		reset();
		props.onClose();
	};

	const handleSubmit = async () => {
		setSubmitted(true);
		if (sourceTab() === "manual" && manualDomains().length === 0) {
			setManualError("Add at least one domain");
			return;
		}
		const source = sourceValue();
		if (!name().trim() || !source.trim()) return;
		try {
			setError("");
			const list = props.list;
			if (list) {
				await updateList(list.id, name(), list.description, source);
			} else {
				await createList(name(), "", source);
			}
			reset();
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
						onChange={(tab) => {
							setSourceTab(tab);
							setManualError("");
						}}
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
				<Show when={sourceTab() === "url"}>
					<TextInput
						label="URL"
						placeholder="https://raw.githubusercontent.com/..."
						hint="Supports hosts files, Adblock Plus filter lists, and plain domain lists."
						value={urlValue()}
						onInput={(e) => setUrlValue(e.currentTarget.value)}
						validate={compose(required(), isUrl())}
						showError={submitted()}
					/>
				</Show>
				<Show when={sourceTab() === "file"}>
					<FileInput
						label="FILE"
						accept=".txt,.csv,.list"
						hint="Supports hosts files, Adblock Plus filter lists, and plain domain lists."
						onChange={setFileValue}
						showError={submitted()}
					/>
				</Show>
				<Show when={sourceTab() === "manual"}>
					<div class="flex flex-col gap-1">
						<div class="flex flex-row items-baseline justify-between gap-4">
							<label
								for="blocklist-domain"
								class="font-medium text-zinc-500 text-sm tracking-wider"
							>
								DOMAINS
							</label>
							<p class="text-sm tracking-wider text-zinc-400">
								{manualDomains().length}
							</p>
						</div>
						<Show when={manualDomains().length > 0}>
							<div class="py-2 flex flex-col">
								<For each={manualDomains()}>
									{(domain) => (
										<div class="py-2 flex flex-row items-center justify-between gap-4 border-b border-zinc-100">
											<span class="min-w-0 truncate text-sm tracking-wider">
												{domain}
											</span>
											<ActionButton
												onclick={() =>
													setManualDomains((domains) =>
														domains.filter((entry) => entry !== domain),
													)
												}
											>
												REMOVE
											</ActionButton>
										</div>
									)}
								</For>
							</div>
						</Show>
						<div class="flex flex-row items-center gap-6">
							<input
								id="blocklist-domain"
								type="text"
								placeholder="example.com"
								value={manualDraft()}
								class="flex-1 min-w-0 py-2 text-sm leading-snug tracking-wider outline-none"
								onInput={(event) => {
									setManualDraft(event.currentTarget.value);
									setManualError("");
								}}
								onKeyDown={(event) => {
									if (event.key === "Enter") {
										event.preventDefault();
										addManualDomain();
									}
								}}
							/>
							<ActionButton onclick={addManualDomain}>ADD</ActionButton>
						</div>
						<Divider />
						<Show when={manualError()}>
							<p class="text-xs text-red-700">{manualError()}</p>
						</Show>
						<p class="text-xs text-zinc-400">
							Press Enter or ADD to add the domain.
						</p>
					</div>
				</Show>
			</div>
			{error() && <p class="text-sm text-red-700">{error()}</p>}
		</Modal>
	);
}
