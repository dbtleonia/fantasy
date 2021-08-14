#!/usr/bin/python3

import bs4
import csv
import email
import sys

out = csv.writer(sys.stdout)
out.writerow(['manager', 'playerid', 'player', 'f_position', 'draftposition', 'lastorank', 'orank'])

with open('lastseason.mhtml') as f:
    for part in email.message_from_file(f).walk():
        if part.get_content_type() == 'text/html':
            soup = bs4.BeautifulSoup(part.get_payload(decode=True), features='html.parser')
            break

main = soup.find('div', {'id': 'yspcontentmainhero'})
for section in main.find_all('section'):
    manager = section.header.h3.get_text()
    for row in section.div.table.tbody.children:
        td = row.find('td', {'class': 'player'})
        playerid = td.a['href'][29:].rstrip('/')
        player = td.a.get_text()
        f_position = td.find('span', {'class': 'F-position'}).get_text()

        draftposition = row.find('td', {'class': 'draftposition'}).get_text()
        lastorank = row.find('td', {'class': 'lastorank'}).get_text()
        orank = row.find('td', {'class': 'orank'}).get_text()

        out.writerow([manager, playerid, player, f_position, draftposition, lastorank, orank])
