#!/usr/bin/python3

import argparse
import bs4
import collections
import csv
import email
import itertools
import os
import re
import sys

def read_managers(indir):
    filename = os.path.join(indir, 'managers.csv')
    print(f'Reading {filename}')
    result = {}
    with open(filename) as f:
        for manager, name in csv.reader(f):
            result[manager] = name
    return result

def read_mhtml(filename):
    print(f'Reading {filename}')
    with open(filename) as f:
        for part in email.message_from_file(f).walk():
            if part.get_content_type() == 'text/html':
                return bs4.BeautifulSoup(part.get_payload(decode=True), features='html.parser')

def read_draft(indir):
    result = {}
    doc = read_mhtml(os.path.join(indir, 'draftresults.mhtml'))
    draft = doc.find('div', {'id': 'drafttables'})
    for drow in draft.find_all('div', {'class': 'Grid-u'}):
        manager = drow.table.thead.tr.th.get_text()
        for prow in drow.table.tbody.find_all('tr'):
            td = prow.find('td', {'class': 'player'})
            pick = int(prow.find('td', {'class': 'pick'}).get_text().lstrip('(').rstrip(')'))
            playerid = td.a['href'][29:].rstrip('/')
            result[playerid] = {
                'round': pick // 12 + 1,
                'kept': td.span is not None,
                # These fields are just for debugging:
                'manager': manager,
                'pick': pick,
                'player': td.a.get_text(),
            }
    return result

def read_dropped(indir):
    result = {}
    for filename in os.listdir(indir):
        if not re.fullmatch(r'dropped\d+\.mhtml', filename):
            continue
        doc = read_mhtml(os.path.join(indir, filename))
        for div in doc.find_all('div', {'class': 'Pbot-xs'}):
            playerid = div.a['href'][29:].rstrip('/')
            player = div.a.get_text()
            result[playerid] = player
    return result

def read_rosters(indir):
    result = []
    for filename in os.listdir(indir):
        if not re.fullmatch(r'roster\d+\.mhtml', filename):
            continue
        doc = read_mhtml(os.path.join(indir, filename))
        manager = str(next(doc
                           .find('div', {'id': 'team-card-info'})
                           .find('div', {'class': 'Ptop-sm'})
                           .find('a')
                           .children)).strip()
        # manager_url = (doc.find('div', {'id': 'team-card-info'})
        #                .find('a', string='View Profile')['href'].split('?')[0])
        player_count = 0
        for i in range(3):
            for tr in doc.find('table', {'id': f'statTable{i}'}).find('tbody').find_all('tr'):
                a = tr.find('a', {'class': 'name'})
                if a is None:
                    continue
                playerid = a['href'][29:].rstrip('/')
                player = a.get_text()
                result.append({
                    'manager': manager,
                    'playerid': playerid,
                    'player': player,
                })
                player_count += 1
        if player_count != 16:
            raise Exception(f'Manager "{manager}" has {player_count} players, want 16')
    return result

UNKEEPABLE = 99  # use large value to sort non-keepables last

def compute_keeper_round(playerid, draft_round,
                         kept1, kept2, kept3,
                         dropped1, dropped2, dropped3):
    # 'teams' in playerid means defense; exempt from 3-year rule
    if (kept1 and kept2 and kept3 and
        not dropped1 and not dropped2 and not dropped3 and
        not 'teams' in playerid):
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
    return (UNKEEPABLE, 'drafted rd 1')

def main(args):
    data_dir = args.data or os.path.join(os.getenv('HOME'), 'data')
    yahoo = os.path.join(data_dir, 'yahoo')

    manager_names = read_managers(os.path.join(yahoo, str(args.year-1)))
    rosters = read_rosters(os.path.join(yahoo, str(args.year-1)))
    draft = {}
    dropped = {}
    for y in range(args.year-1, args.year-4, -1):
        draft[y] = read_draft(os.path.join(yahoo, str(y)))
        dropped[y] = read_dropped(os.path.join(yahoo, str(y)))

    result = []
    for p in rosters:
        manager_name = manager_names[p['manager']]
        playerid = p['playerid']
        if playerid == 'teams/la-chargers':
            player = 'Los Angeles Chargers'
        elif playerid == 'teams/la-rams':
            player = 'Los Angeles Rams'
        elif playerid == 'teams/ny-giants':
            player = 'New York Giants'
        elif playerid == 'teams/ny-jets':
            player = 'New York Jets'
        else:
            player = p['player']
        draft_round = draft[args.year-1][playerid]['round'] if playerid in draft[args.year-1] else None
        kept1 = playerid in draft[args.year-1] and draft[args.year-1][playerid]['kept']
        kept2 = playerid in draft[args.year-2] and draft[args.year-2][playerid]['kept']
        kept3 = playerid in draft[args.year-3] and draft[args.year-3][playerid]['kept']
        dropped1 = playerid in dropped[args.year-1]
        dropped2 = playerid in dropped[args.year-2]
        dropped3 = playerid in dropped[args.year-3]
        keeper_round, reason = compute_keeper_round(
            playerid, draft_round,
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

    d = os.path.join(data_dir, 'out', str(args.year))
    os.makedirs(d, exist_ok=True)
    admin_csv = csv.writer(open(os.path.join(d, 'keeper_options.csv'), 'w'))
    admin_csv.writerow([
        'Manager',
        'Player',
        'Player ID',
        f'Keeper round {args.year % 100}',
        f'Draft round {(args.year-1) % 100}',
        'Reason',
        f'Kept {(args.year-1) % 100}',
        f'Dropped {(args.year-1) % 100}',
        f'Kept {(args.year-2) % 100}',
        f'Dropped {(args.year-2) % 100}',
        f'Kept {(args.year-3) % 100}',
        f'Dropped {(args.year-3) % 100}',
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
                k1,
                d1,
                k2,
                d2,
                k3,
                d3,
            ])

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('year', type=int, help='the year to generate files for')
    parser.add_argument('--data', help='directory for data files, defaults to $HOME/data')
    args = parser.parse_args()
    main(args)
