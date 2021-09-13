#!/bin/sh

ori=$1
if [ -z "$ori" ]; then
    ori=origin
fi

flocal=git-slow-local-hash-long-so-dont-conflikct.txt
ftag=git-slow-remote-tag-long-so-dont-conflikct.txt
fremote=git-slow-remote-hash-long-so-dont-conflikct.txt

if [ -e $flocal ]; then
    rm $flocal
fi

#echo "read [local] branch to $flocal"
git branch -av | grep "remotes\/${ori}\/" | sed "s/remotes\/${ori}\//local /g" | awk '{print $1" "$3" "$2}' > $flocal
#echo "read [local] tag to $flocal"
git tag --list --format="local %(objectname) %(refname:short)" >> $flocal
echo "[local] branch&tag count: `cat $flocal | wc -l`"

if [ -e $ftag ]; then
    rm $ftag
fi

#echo "get [remote $ori] tag to $ftag"
git ls-remote --tags --sort='v:refname' $ori | grep "refs\/tags\/" | sed 's/refs\/tags\///g' | awk '{print "tag "$1" "$2}' > $ftag
echo "[remote $ori] tag count: `cat $ftag | wc -l`"

if [ -e $fremote ]; then
    rm $fremote
fi

#echo "get [remote $ori] branch to $fremote"
git ls-remote $ori | grep "refs\/heads\/" | sed 's/refs\/heads\///g' | awk '{print "remote "$1" "$2}' > $fremote
echo "[remote $ori] branch count: `cat $fremote | wc -l`"

echo

cat $flocal $ftag $fremote | awk 'BEGIN{
    split("", b);
    split("", e);
    split("", t);
    state = "";
}{
    if ($1 != stage) {
        stage = $1;
        if (stage == "local") {
            printf("start counting local branch&tag...\n");
        } else if (stage == "tag") {
            printf("start matching remote tag...\n");
        } else if (stage == "remote") {
            printf("start matching remote branch...\n");
        }
    }

    if ($1 == "local") {
        a[$3]=$2;
    }

    if ($1 == "tag") {
        if (index($3, "^{}") == 0) {
            h = a[$3];
            if (length(h) == 0 || index($2, h) != 1) {
                if (0 != system("git log -1 --oneline "$3" > /dev/null 2>&1")) {
                    t[$3] = $2;
                }
            }
        }
    }

    if ($1 == "remote") {
        h = a[$3];
        if (length(h) == 0 || index($2, h) != 1) {
            if (0 != system("git log -1 --oneline "$2" > /dev/null 2>&1")) {
                b[$3] = $2;
                c++;
            } else {
                e[$3] = $2;
            }
        }
    }
}END{
    if (length(t) > 0) {
        printf("\n%d new tag, fetch...\n", length(t));
        for (n in t) {
            tt++;
            printf("\n[%d/%d] fetch tag %s\n", tt, length(t), n);
            system("git fetch '${ori}' "n);
        }
    }

    if (c > 1) {
        printf("\n%d new branches, fetch...\n", c);
    } else if (c == 1) {
        printf("\none new branch, fetch...\n");
    }

    for (n in b) {
        cc++;
        printf("\n[%d/%d] fetch %s %s\n", cc, c, b[n], n);
        system("git fetch '${ori}' "n);
    }

    if (length(e) > 0) {
        printf("\n%s branches updated, but hashes exist, try fetch:\n", length(e));
        system("git fetch '${ori}'");
    }

    printf("\nall branches and tags have been fetched\n");
}'

rm $flocal
rm $ftag
rm $fremote

exit;

echo "run(git remote show $ori)..."

git remote show $ori | awk '{
    if ($2 == "new" || $2=="新的") {
        c++;
        b[$1]=c;
    }

    if ($2 == "trackered") {
        old++;
    }
} END {
    print "new branch count: "c;
    print "old branch count: "old;
    for (n in b) {
        printf("[%3d/%3d] fetch %s\n", b[n], c, n);
        system("git fetch origin "n);
    }
}'
