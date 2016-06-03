#!/usr/bin/awk -f
BEGIN {
    FS = "^"
    OFS = ""
    for (i = 1; i < ARGC; i++) {
	    if (ARGV[i] == "--test")
	    {
	        test = 1
		    delete ARGV[i]
		    break
		}
    }
}
{
	if (test == 1) {
		if ($1 == "{test}") {
			$1 = ""			
			print
		}
		else if ($1 != "{main}") {
			print
		}
	}
	else
	{
		if ($1 == "{main}") {
			$1 = ""			
			print 
		}
		else if ($1 != "{test}") {
			print
		}
	}
}
