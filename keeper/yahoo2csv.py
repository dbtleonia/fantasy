#!/usr/bin/python3

import argparse
import bs4
import csv
import email
import itertools
import os
import re
import sys
from collections import defaultdict

def read_managers(indir):
    result = {}
    with open(os.path.join(indir, 'managers.csv')) as f:
        for manager, name in csv.reader(f):
            result[manager] = name
    return result

def read_mhtml(filename):
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
            playerid = td.a['href'][29:].rstrip('/')
            result[playerid] = {
                'round': int(prow.find('td', {'class': 'first'}).get_text().rstrip('.')),
                'kept': td.span is not None,
                # These fields are just for debugging:
                'manager': manager,
                'pick': int(prow.find('td', {'class': 'pick'}).get_text().lstrip('(').rstrip(')')),
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

def keeper_round(dround, kept, dropt):
    if dround is None:
        return 11 if dropt else 9
    if dropt:
        return min(16, dround + 2)
    if not kept:
        return dround
    if dround > 1:
        return dround - 1
    # Return large value to sort non-keepables last.
    return 99

def main(args):
    data_dir = args.data
    if data_dir is None:
        data_dir = os.path.join(os.getenv('HOME'), 'kdata')
    yahoo = os.path.join(data_dir, 'yahoo')

    manager_names = read_managers(data_dir)
    draft = read_draft(yahoo)
    dropped = read_dropped(yahoo)
    rosters = read_rosters(yahoo)

    result = []
    for p in rosters:
        manager_name = manager_names[p['manager']]
        playerid = p['playerid']
        player = p['player']
        dround = draft[playerid]['round'] if playerid in draft else None
        kept = playerid in draft and draft[playerid]['kept']
        dropt = playerid in dropped
        kround = keeper_round(dround, kept, dropt)
        k = 'K' if kept else ''
        d = 'D' if dropt else ''
        result.append([manager_name, kround, player, playerid, dround, k, d])
    result.sort()

    os.makedirs(os.path.join(data_dir, 'toimport', 'options'), exist_ok=True)
    admin_csv = csv.writer(open(os.path.join(data_dir, 'toimport', 'admin.csv'), 'w'))
    admin_csv.writerow(['manager', 'player', 'playerid', 'dround', 'kept', 'dropt', 'kround'])

    for manager_name, group in itertools.groupby(result, lambda r: r[0]):
        manager_csv = csv.writer(open(os.path.join(data_dir, 'toimport', 'options', manager_name), 'w'))
        manager_csv.writerow(['Player', 'Player ID', 'Keeper Round'])

        for _, kround, player, playerid, dround, k, d in group:
            if kround == 99:
                kround = 'n/a'
            manager_csv.writerow([player, playerid, kround])
            admin_csv.writerow([manager_name, player, playerid, dround, k, d, kround])

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--data', help='directory for data files, defaults to $HOME/kdata')
    args = parser.parse_args()
    main(args)
