package commands

import "testing"

func TestContentDetection(t *testing.T) {
	matches := []string{
		"content",
		"Content",
		"CONTENT",
		"some content here",
		"c0ntent",
		"c0nt3nt",
		"c0n73n7",
		"c2ontent",
		"cqntent",
		"cQntent",
		"сontent",
		"cоntent",
		"cоntеnt",
		"сонтент",
		"СONTENT",
		"cОntent",
		"cοntent",
		"ϲontent",
		"ｃontent",
		"ｃｏｎｔｅｎｔ",
		"c-o-n-t-e-n-t",
		"c o n t e n t",
		"c_o_n_t_e_n_t",
		"c·o·n·t·e·n·t",
		"c​ontent",
		"cöntent",
		"cóntent",
		"c̶o̶n̶t̶e̶n̶t",
		"ᴄontent",
		"ᴄᴏɴᴛᴇɴᴛ",
		"𝐜𝐨𝐧𝐭𝐞𝐧𝐭",
		"ⓒⓞⓝⓣⓔⓝⓣ",
		"coⁿtent",
		"[ontent",
		"(ontent",
		"{ontent",
		"（ontent",
		"[0ntent",
		"[(ontent",
		"co[n-t-e-n-t",
		"ʇuǝʇuoɔ",
		"ʇuǝʇuoɔ!",
		"ʇ u ǝ ʇ u o ɔ",
		"tnetnoc",
		"⊥NƎ⊥NOƆ",
		"contǝnt",
		"conʇent",
	}

	nonMatches := []string{
		"",
		"hi",
		"69",
		"context",
		"continent",
		"consent",
		"connect",
		"contest",
		"constant",
		"continuous",
		"contact",
		"concert",
		"concentrate",
		"c o n t a i n",
		"cooking on the table",
		"ice cream on Tuesday",
		"pictures of unicorns",
		"chicken nuggets",
		"nice one",
		"(on top of the list)",
		"(oof)",
		"[okay then]",
		"nice one on the internet, eat?",
		"check out on top of the list",
		"unique question, nice",
	}

	for _, s := range matches {
		if !containsContent(s) {
			t.Errorf("expected %q to match content", s)
		}
	}

	for _, s := range nonMatches {
		if containsContent(s) {
			t.Errorf("expected %q NOT to match content", s)
		}
	}
}
