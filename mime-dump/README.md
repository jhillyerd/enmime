mime-dump
=========

mime-dump is a utility that aids in the debugging of go.enmime.  To use it type
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
<html><head></head><body style="word-wrap: break-word; -webkit-nbsp-mode: space; -webkit-line-break: after-white-space; "><font class="Apple-style-span" face="'Comic Sans MS'">Test of HTML section</font><img height="16" width="16" apple-width="yes" apple-height="yes" id="4579722f-d53d-45d0-88bc-f8209a2ca569" src="cid:8B8481A2-25CA-4886-9B5A-8EB9115DD064@skynet"></body></html>

Attachment List
---------------

MIME Part Tree
--------------
    multipart/alternative
    |-- text/plain
    `-- multipart/related
        |-- text/html
        `-- image/png, disposition: inline, filename: "favicon.png"
