
/* RescueTime::Network::API::activate_silent(std::__cxx11::basic_string<wchar_t,
   std::char_traits<wchar_t>, std::allocator<wchar_t> > const&) */

API * __thiscall RescueTime::Network::API::activate_silent(API *this,basic_string *param_1)

{
  size_t sVar1;
  long in_FS_OFFSET;
  undefined *local_108 [2];
  undefined auStack_f8 [16];
  undefined *local_e8 [2];
  undefined local_d8 [16];
  undefined *local_c8 [2];
  undefined auStack_b8 [16];
  undefined *local_a8 [2];
  undefined auStack_98 [16];
  wchar_t *local_88;
  int local_80;
  wchar_t local_78 [4];
  undefined *local_68;
  undefined local_58 [24];
  long local_40;

  local_40 = *(long *)(in_FS_OFFSET + 0x28);
  std::__cxx11::basic_string<>::basic_string
            ((wchar_t *)&local_88,(allocator *)L"[API::activate_silent]");
                    /* try { // try from 00631398 to 0063139c has its CatchHandler @ 006315c3 */
  rt::debug_log((basic_string *)&local_88,0xc0);
  if (local_88 != local_78) {
    operator.delete(local_88);
  }
  std::__cxx11::basic_string<>::basic_string((wchar_t *)local_a8,(allocator *)&DAT_009f06ec);
  local_c8[0] = auStack_b8;
  sVar1 = wcslen(L"");
                    /* try { // try from 00631404 to 00631408 has its CatchHandler @ 006315be */
  std::__cxx11::basic_string<>::_M_construct<>((wchar_t *)local_c8,L"",(int)sVar1 * 4 + 0x9f06ec);
  local_e8[0] = local_d8;
  sVar1 = wcslen(L"");
                    /* try { // try from 00631437 to 0063143b has its CatchHandler @ 006315b9 */
  std::__cxx11::basic_string<>::_M_construct<>((wchar_t *)local_e8,L"",(int)sVar1 * 4 + 0x9f06ec);
                    /* try { // try from 0063144f to 00631453 has its CatchHandler @ 006315b4 */
  std::__cxx11::basic_string<>::basic_string((wchar_t *)local_108,(allocator *)&DAT_009f06ec);
                    /* try { // try from 0063146e to 00631472 has its CatchHandler @ 0063154e */
  perform_activate((basic_string *)&local_88,param_1,(basic_string *)local_108,
                   (basic_string *)local_e8);
  if (local_108[0] != auStack_f8) {
    operator.delete(local_108[0]);
  }
  if (local_e8[0] != local_d8) {
    operator.delete(local_e8[0]);
  }
  if (local_c8[0] != auStack_b8) {
    operator.delete(local_c8[0]);
  }
  if (local_a8[0] != auStack_98) {
    operator.delete(local_a8[0]);
  }
  *(API **)this = this + 0x10;
                    /* try { // try from 006314ea to 006314ee has its CatchHandler @ 0063153b */
  std::__cxx11::basic_string<>::_M_construct<>
            ((wchar_t *)this,local_88,(int)local_88 + local_80 * 4);
  if (local_68 != local_58) {
    operator.delete(local_68);
  }
  if (local_88 != local_78) {
    operator.delete(local_88);
  }
  if (local_40 == *(long *)(in_FS_OFFSET + 0x28)) {
    return this;
  }
                    /* WARNING: Subroutine does not return */
  __stack_chk_fail();
}

perform_activate