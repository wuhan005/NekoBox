export interface ChangeLogItem {
    content: string;
    date: string;
}

export async function getChangeLogs() {
    return await (await fetch('https://json.n3ko.cc/nekobox-updatehistory/index.json')).json()
}

export async function getSponsors() {
    return await (await fetch('https://json.n3ko.cc/nekobox-sponsor/index.json')).json()
}
