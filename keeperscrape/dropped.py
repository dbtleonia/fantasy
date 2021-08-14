#!/usr/bin/python3

import bs4
import csv
import email
import sys

out = csv.writer(sys.stdout)
out.writerow(['playerid', 'player'])

for n in range(0, 250, 25):
    with open(f'dropped{n:03d}.mhtml') as f:
        for part in email.message_from_file(f).walk():
            if part.get_content_type() == 'text/html':
                soup = bs4.BeautifulSoup(part.get_payload(decode=True), features='html.parser')
                break

    for div in soup.find_all('div', {'class': 'Pbot-xs'}):
        playerid = div.a['href'][29:].rstrip('/')
        player = div.a.get_text()
        out.writerow([playerid, player])
