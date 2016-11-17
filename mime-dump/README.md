mime-dump
=========

mime-dump is a utility that aids in the debugging of enmime.  To use it type
`go build` in this directory, then pass it an email to parse:

    ./mime-dump ../test-data/mail/html-mime-inline.raw

If all goes well, it will output a markdown formatted document describing the
email...

----

html-mime-inline.raw
====================

Envelope
--------
From: James Hillyerd <james@makita.skynet>  
To: greg@nobody.com  
Subject: MIME test 1  

Body Text
---------
Test of text section

Body HTML
---------
Test of HTML section

Attachment List
---------------

MIME Part Tree
--------------
    multipart/alternative
    |-- text/plain
    `-- multipart/related
        |-- text/html
        `-- image/png, disposition: inline, filename: "favicon.png"
