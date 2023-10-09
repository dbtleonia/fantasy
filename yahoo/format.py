#!/usr/bin/python3

import csv
import json
import os
import sys
from operator import *


def home(path):
    return os.path.join(os.path.expanduser('~'), path)


def league_scoreboard(week):
    with open(home(f'league_scoreboard_week{week:02d}.json')) as f:
        data = json.load(f)

    out = home(f'league_scoreboard_week{week:02d}.csv')
    with open(out, 'w') as f:
        w = csv.writer(f)
        w.writerow([
            'team1',
            'team1 points',
            'team1 projected points',
            'team2',
            'team2 points',
            'team2 projected points'])
        for m in data['fantasy_content']['league']['scoreboard']['matchups']:
            row = []
            for t in m['matchup']['teams']:
                row.extend([
                    t['team']['name'],
                    t['team']['team_points']['total'],
                    t['team']['team_projected_points']['total']])
            w.writerow(row)
    print(f'Wrote {out}')


def league_standings(suffix):
    with open(home(f'league_settings_{suffix}.json')) as f:
        settings = json.load(f)
    with open(home(f'league_standings_{suffix}.json')) as f:
        standings = json.load(f)

    divisions = {}  # id -> name
    for d in settings['fantasy_content']['league']['settings']['divisions']:
        divisions[d['division']['division_id']] = d['division']['name']

    teams = []
    for t in standings['fantasy_content']['league']['standings']['teams']:
        teams.append(t['team'])
    teams.sort(key=itemgetter('division_id'))

    out = home(f'league_standings_{suffix}.csv')
    with open(out, 'w') as f:
        w = csv.writer(f)
        w.writerow([
            'division',
            'team',
            'points',
            'wins',
            'losses',
            'ties'])
        for team in teams:
            w.writerow([
                divisions[int(team['division_id'])],
                team['name'],
                team['team_points']['total'],
                team['team_standings']['outcome_totals']['wins'],
                team['team_standings']['outcome_totals']['losses'],
                team['team_standings']['outcome_totals']['ties']])
    print(f'Wrote {out}')


def team_matchups():
    data = []
    for t in range(1, 12+1):
        with open(home(f'team_matchups_{t:02d}.json')) as f:
            data.append(json.load(f))

    out = home('team_matchups.csv')
    with open(out, 'w') as f:
        w = csv.writer(f)
        w.writerow([
            'team',
            'week',
            'opponent'])
        for t in data:
            for m in t['fantasy_content']['team']['matchups']:
                w.writerow([
                    m['matchup']['teams'][0]['team']['name'],
                    m['matchup']['week'],
                    m['matchup']['teams'][1]['team']['name']])
    print(f'Wrote {out}')


def team_rosters(week):
    data = []
    for t in range(1, 12+1):
        with open(home(f'team_rosters_week{week:02d}_{t:02d}.json')) as f:
            data.append(json.load(f))

    out = home(f'team_rosters_week{week:02d}.csv')
    with open(out, 'w') as f:
        w = csv.writer(f)
        w.writerow([
            'team',
            'player',
            'position',
            'points'])
        for t in data:
            team = t['fantasy_content']['team']
            for p in team['roster']['players']:
                w.writerow([
                    team['name'],
                    p['player']['name']['full'],
                    p['player']['selected_position']['position'],
                    p['player']['player_points']['total']])
    print(f'Wrote {out}')


def league_players_available(suffix):
    # TODO: Don't hardcode limit.
    players = []
    for start in range(0, 950, 25):
        with open(home(f'league_players_available_{suffix}_{start:04d}.json')) as f:
            data = json.load(f)
            players.extend(data['fantasy_content']['league']['players'])

    out = home(f'league_players_available_{suffix}.csv')
    with open(out, 'w') as f:
        w = csv.writer(f)
        w.writerow([
            'player',
            'position',
            'points'])
        for p in players:
            w.writerow([
                p['player']['name']['full'],
                p['player']['primary_position'],
                p['player']['player_points']['total']])
    print(f'Wrote {out}')


def main(argv):
    if len(argv) < 2:
        sys.exit('gimme cmd')
    if argv[1] == 'a':
        if len(argv) != 3:
            sys.exit('gimme suffix')
        league_players_available(argv[2])
    elif argv[1] == 'b':
        if len(argv) != 3:
            sys.exit('gimme week')
        league_scoreboard(int(argv[2]))
    elif argv[1] == 'm':
        team_matchups()
    elif argv[1] == 'r':
        if len(argv) != 3:
            sys.exit('gimme week')
        team_rosters(int(argv[2]))
    elif argv[1] == 's':
        if len(argv) != 3:
            sys.exit('gimme suffix')
        league_standings(argv[2])
    else:
        sys.exit(f'unknown command {argv[1]}')


if __name__ == '__main__':
    main(sys.argv)
