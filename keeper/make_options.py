#!/usr/bin/python3

import argparse
import bs4
import collections
import csv
import email
import itertools
import json
import os
import re
import sys

def read_managers(indir, year):
    data = read_json(os.path.join(indir, f'managers-{year}.json'))
    result = {}
    for t in data['fantasy_content']['league']['teams']:
        manager = t['team']['managers'][0]['manager']
        result[manager['guid']] = f"{manager['nickname']} - {t['team']['name']}"
    return result

def read_json(filename):
    print(f'Reading {filename}')
    with open(filename) as f:
        return json.load(f)

def read_mhtml(filename):
    print(f'Reading {filename}')
    with open(filename) as f:
        for part in email.message_from_file(f).walk():
            if part.get_content_type() == 'text/html':
                return bs4.BeautifulSoup(part.get_payload(decode=True), features='html.parser')

def read_draft(indir, year, url_ids):
    result = {}
    doc = read_mhtml(os.path.join(indir, f'draftresults-{year}.mhtml'))
    draft = doc.find('div', {'id': 'drafttables'})
    for drow in draft.find_all('div', {'class': 'Grid-u'}):
        manager = drow.table.thead.tr.th.get_text()
        for prow in drow.table.tbody.find_all('tr'):
            td = prow.find('td', {'class': 'player'})
            pick = int(prow.find('td', {'class': 'pick'}).get_text().lstrip('(').rstrip(')'))
            if td.a['href'] not in url_ids:
                continue
            playerid = url_ids[td.a['href']]
            result[playerid] = {
                'round': pick // 12 + 1,
                'kept': td.span is not None,
                # These fields are just for debugging:
                'manager': manager,
                'pick': pick,
                'player': td.a.get_text(),
            }
    return result

def read_dropped(indir, year1, year2, year3):
    data = read_json(os.path.join(indir, f'drops-{year1}-{year2}-{year3}.json'))
    result = {}
    for l in data['fantasy_content']['leagues']:
        season = int(l['league']['season'])
        result[season] = {}
        for t in l['league']['transactions']:
            for p in t['transaction']['players']:
                if p['player']['transaction_data']['type'] == 'drop':
                    playerid = p['player']['player_id']
                    result[season][playerid] = p['player']['name']['full']
    return result

def read_rosters(indir, year):
    data = read_json(os.path.join(indir, f'rosters-{year}.json'))
    result = []
    url_ids = {}
    for t in data['fantasy_content']['league']['teams']:
        manager = t['team']['managers'][0]['manager']
        for p in t['team']['players']:
            result.append({
                'manager': manager['guid'],
                'playerid': p['player']['player_id'],
                'player': f"{p['player']['name']['full']} ({p['player']['editorial_team_abbr']} - {p['player']['display_position']})",
                'position': p['player']['primary_position']
            })
            url_ids[p['player']['url']] = p['player']['player_id']
    return result, url_ids

UNKEEPABLE = 99  # use large value to sort non-keepables last

def compute_keeper_round(playerid, position, draft_round,
                         kept1, kept2, kept3,
                         dropped1, dropped2, dropped3):
    if (kept1 and kept2 and kept3 and
        not dropped1 and not dropped2 and not dropped3 and
        position != 'DEF'):
        return (UNKEEPABLE, 'kept 3 years non-D')
    if draft_round is None:
        if dropped1:
            return (11, 'undrafted/dropped')
        else:
            return (9, 'undrafted')
    if dropped1:
        # If the player would have been unkeepable but was dropped,
        # flag for manual check.
        if ((kept1 and kept2 and kept3 and
             not dropped2 and not dropped3 and
             not 'teams' in playerid) or
            draft_round == 1):
            return (min(16, draft_round + 2), 'dropped/saquon')
        else:
            return (min(16, draft_round + 2), 'dropped')
    if not kept1:
        return (draft_round, '')
    if draft_round > 1:
        return (draft_round - 1, 'kept')
    return (UNKEEPABLE, 'kept round 1')

def main(args):
    data_dir = args.data or os.path.join(os.getenv('HOME'), 'data')
    yahoo = os.path.join(data_dir, 'yahoo')

    # From json files.
    manager_names = read_managers(yahoo, args.year)      # {guid: name}
    rosters, url_ids = read_rosters(yahoo, args.year-1)  # [{'manager': guid, 'playerid': id, 'player': name, 'position': pos}], {url: id}
    dropped = read_dropped(yahoo, args.year-1, args.year-2, args.year-3)  # {season: {id: name}}

    # From mhtml files because json is buggy.
    draft = {}  # {season: {id: {'round': <int>, 'kept': <bool>}}}
    for y in range(args.year-1, args.year-4, -1):
        draft[y] = read_draft(yahoo, y, url_ids)

    result = []
    for p in rosters:
        manager_name = manager_names[p['manager']]
        playerid = p['playerid']
        player = p['player']
        position = p['position']
        draft_round = draft[args.year-1][playerid]['round'] if playerid in draft[args.year-1] else None
        kept1 = playerid in draft[args.year-1] and draft[args.year-1][playerid]['kept']
        kept2 = playerid in draft[args.year-2] and draft[args.year-2][playerid]['kept']
        kept3 = playerid in draft[args.year-3] and draft[args.year-3][playerid]['kept']
        dropped1 = playerid in dropped[args.year-1]
        dropped2 = playerid in dropped[args.year-2]
        dropped3 = playerid in dropped[args.year-3]
        keeper_round, reason = compute_keeper_round(
            playerid, position, draft_round,
            kept1, kept2, kept3,
            dropped1, dropped2, dropped3)

        result.append([
            manager_name,
            keeper_round,
            reason,
            player,
            playerid,
            draft_round,
            f'K{(args.year-1) % 100}' if kept1 else '',
            f'K{(args.year-2) % 100}' if kept2 else '',
            f'K{(args.year-3) % 100}' if kept3 else '',
            f'D{(args.year-1) % 100}' if dropped1 else '',
            f'D{(args.year-2) % 100}' if dropped2 else '',
            f'D{(args.year-3) % 100}' if dropped3 else '',
        ])

    for player, n in collections.Counter(p[3] for p in result).items():
        if n > 1:
            sys.exit(f'*** ERROR *** Duplicate name: {player}')

    result.sort()

    d = os.path.join(data_dir, 'out')
    os.makedirs(d, exist_ok=True)
    filename = os.path.join(d, f'keeper-options.csv')
    print(f'Writing {filename}')
    admin_csv = csv.writer(open(filename, 'w'))
    admin_csv.writerow([
        'Manager',
        'Player',
        'Player ID',
        f'Keeper round {args.year % 100}',
        f'Draft round {(args.year-1) % 100}',
        'Reason',
        f'Dropped {(args.year-1) % 100}',
        f'Kept {(args.year-1) % 100}',
        f'Dropped {(args.year-2) % 100}',
        f'Kept {(args.year-2) % 100}',
        f'Dropped {(args.year-3) % 100}',
        f'Kept {(args.year-3) % 100}',
    ])
    for manager_name, group in itertools.groupby(result, lambda r: r[0]):
        for _, keeper_round, reason, player, playerid, draft_round, k1, k2, k3, d1, d2, d3, in group:
            if keeper_round == UNKEEPABLE:
                keeper_round = 'n/a'
            admin_csv.writerow([
                manager_name,
                player,
                playerid,
                keeper_round,
                draft_round,
                reason,
                d1,
                k1,
                d2,
                k2,
                d3,
                k3,
            ])

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('year', type=int, help='the year to generate files for')
    parser.add_argument('--data', help='directory for data files, defaults to $HOME/data')
    args = parser.parse_args()
    main(args)
